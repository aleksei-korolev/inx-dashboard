package dashboard

import (
	"github.com/labstack/echo/v4"

	"github.com/iotaledger/hive.go/core/syncutils"
	"github.com/iotaledger/hive.go/core/websockethub"
	"github.com/iotaledger/inx-dashboard/pkg/jwt"
)

const (
	// MsgTypeSyncStatus is the type of the SyncStatus message.
	MsgTypeSyncStatus byte = iota
	// MsgTypePublicNodeStatus is the type of the PublicNodeStatus message.
	MsgTypePublicNodeStatus = 1
	// MsgTypeNodeInfoExtended is the type of the NodeInfoExtended message.
	MsgTypeNodeInfoExtended = 2
	// MsgTypeGossipMetrics is the type of the GossipMetrics message.
	MsgTypeGossipMetrics = 3
	// MsgTypeMilestone is the type of the Milestone message.
	MsgTypeMilestone = 4
	// MsgTypePeerMetric is the type of the PeerMetric message.
	MsgTypePeerMetric = 5
	// MsgTypeConfirmedMsMetrics is the type of the ConfirmedMsMetrics message.
	MsgTypeConfirmedMsMetrics = 6
	// MsgTypeVisualizerVertex is the type of the Vertex message for the visualizer.
	MsgTypeVisualizerVertex = 7
	// MsgTypeVisualizerSolidInfo is the type of the SolidInfo message for the visualizer.
	MsgTypeVisualizerSolidInfo = 8
	// MsgTypeVisualizerConfirmedInfo is the type of the ConfirmedInfo message for the visualizer.
	MsgTypeVisualizerConfirmedInfo = 9
	// MsgTypeVisualizerMilestoneInfo is the type of the MilestoneInfo message for the visualizer.
	MsgTypeVisualizerMilestoneInfo = 10
	// MsgTypeVisualizerTipInfo is the type of the TipInfo message for the visualizer.
	MsgTypeVisualizerTipInfo = 11
	// MsgTypeDatabaseSizeMetric is the type of the database Size message for the metrics.
	MsgTypeDatabaseSizeMetric = 12
)

func (d *Dashboard) websocketRoute(ctx echo.Context) error {
	defer func() {
		if r := recover(); r != nil {
			d.LogErrorf("recovered from panic within WS handle func: %s", r)
		}
	}()

	publicTopics := []byte{
		MsgTypeSyncStatus,
		MsgTypePublicNodeStatus,
		MsgTypeGossipMetrics,
		MsgTypeMilestone,
		MsgTypeConfirmedMsMetrics,
		MsgTypeVisualizerVertex,
		MsgTypeVisualizerSolidInfo,
		MsgTypeVisualizerConfirmedInfo,
		MsgTypeVisualizerMilestoneInfo,
		MsgTypeVisualizerTipInfo,
	}

	isProtectedTopic := func(topic byte) bool {
		for _, publicTopic := range publicTopics {
			if topic == publicTopic {
				return false
			}
		}

		return true
	}

	// this function sends the initial values for some topics
	sendInitValue := func(client *websockethub.Client, initValuesSent map[byte]struct{}, topic byte) {
		// always send the initial values for the Vertex topic, ignore others that were already sent
		if _, sent := initValuesSent[topic]; sent && (topic != MsgTypeVisualizerVertex) {
			return
		}
		initValuesSent[topic] = struct{}{}

		switch topic {

		case MsgTypeSyncStatus:
			client.Send(&Msg{Type: MsgTypeSyncStatus, Data: d.getSyncStatus()})

		case MsgTypePublicNodeStatus:
			nodeInfo, err := d.getNodeInfo()
			if err != nil {
				d.LogWarnf("failed to get node info: %s", err)

				return
			}

			data := getPublicNodeStatusByNodeInfo(nodeInfo, d.nodeBridge.IsNodeAlmostSynced())
			d.hub.BroadcastMsg(&Msg{Type: MsgTypePublicNodeStatus, Data: data})

		case MsgTypeNodeInfoExtended:
			data, err := d.getNodeInfoExtended()
			if err != nil {
				d.LogWarnf("failed to get extended node info: %s", err)

				return
			}
			client.Send(&Msg{Type: MsgTypeNodeInfoExtended, Data: data})

		case MsgTypeGossipMetrics:
			data, err := d.getGossipMetrics()
			if err != nil {
				d.LogWarnf("failed to get gossip metrics: %s", err)

				return
			}
			client.Send(&Msg{Type: MsgTypeGossipMetrics, Data: data})

		case MsgTypeMilestone:
			start := d.getLatestMilestoneIndex()
			for msIndex := start - 10; msIndex <= start; msIndex++ {
				if milestoneIDHex, err := d.getMilestoneIDHex(msIndex); err == nil {
					client.Send(&Msg{Type: MsgTypeMilestone, Data: &Milestone{MilestoneID: milestoneIDHex, Index: msIndex}})
				} else {
					d.LogWarnf("failed to get milestone %d: %s", msIndex, err)

					return
				}
			}

		case MsgTypePeerMetric:
			data, err := d.getPeerInfos()
			if err != nil {
				d.LogWarnf("failed to get peer infos: %s", err)

				return
			}
			client.Send(&Msg{Type: MsgTypePeerMetric, Data: data})

		case MsgTypeConfirmedMsMetrics:
			data, err := d.getNodeInfo()
			if err != nil {
				d.LogWarnf("failed to get node info: %s", err)

				return
			}
			client.Send(&Msg{Type: MsgTypeConfirmedMsMetrics, Data: data.Metrics})

		case MsgTypeVisualizerVertex:
			d.visualizer.ForEachCreated(func(vertex *VisualizerVertex) bool {
				// don't drop the messages to fill the visualizer without missing any vertex
				client.Send(&Msg{Type: MsgTypeVisualizerVertex, Data: vertex}, true)

				return true
			}, VisualizerInitValuesCount)

		case MsgTypeDatabaseSizeMetric:
			client.Send(&Msg{Type: MsgTypeDatabaseSizeMetric, Data: d.cachedDatabaseSizeMetrics})
		}
	}

	topicsLock := syncutils.RWMutex{}
	registeredTopics := make(map[byte]struct{})
	initValuesSent := make(map[byte]struct{})

	d.hub.ServeWebsocket(ctx.Response(), ctx.Request(),
		// onCreate gets called when the client is created
		func(client *websockethub.Client) {
			client.FilterCallback = func(_ *websockethub.Client, data interface{}) bool {
				msg, ok := data.(*Msg)
				if !ok {
					return false
				}

				topicsLock.RLock()
				_, registered := registeredTopics[msg.Type]
				topicsLock.RUnlock()

				return registered
			}
			client.ReceiveChan = make(chan *websockethub.WebsocketMsg, 100)

			go func() {
				for {
					select {
					case <-client.ExitSignal:
						// client was disconnected
						return

					case msg, ok := <-client.ReceiveChan:
						if !ok {
							// client was disconnected
							return
						}

						if msg.MsgType == websockethub.BinaryMessage {
							if len(msg.Data) < 2 {
								continue
							}

							cmd := msg.Data[0]
							topic := msg.Data[1]

							if cmd == WebsocketCmdRegister {

								if isProtectedTopic(topic) {
									// Check for the presence of a JWT and verify it
									if len(msg.Data) < 3 {
										// Dot not allow unsecure subscriptions to protected topics
										continue
									}
									token := string(msg.Data[2:])
									if !d.jwtAuth.VerifyJWT(token, func(claims *jwt.AuthClaims) bool {
										return true
									}) {
										// Dot not allow unsecure subscriptions to protected topics
										continue
									}
								}

								// register topic fo this client
								topicsLock.Lock()
								registeredTopics[topic] = struct{}{}
								topicsLock.Unlock()

								sendInitValue(client, initValuesSent, topic)

							} else if cmd == WebsocketCmdUnregister {
								// unregister topic fo this client
								topicsLock.Lock()
								delete(registeredTopics, topic)
								topicsLock.Unlock()
							}
						}
					}
				}
			}()
		},

		// onConnect gets called when the client was registered
		func(_ *websockethub.Client) {
			d.LogInfo("WebSocket client connection established")
		})

	return nil
}
