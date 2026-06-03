package metrics

import "github.com/prometheus/client_golang/prometheus"

const namespace = "nexus"

var (
	// ActiveSessions 当前活跃的客户端 session 数
	ActiveSessions = prometheus.NewGauge(prometheus.GaugeOpts{
		Namespace: namespace,
		Name:      "sessions_active",
		Help:      "当前活跃的客户端 session 数",
	})

	// ActiveTunnels 当前活跃的隧道监听器数
	ActiveTunnels = prometheus.NewGauge(prometheus.GaugeOpts{
		Namespace: namespace,
		Name:      "tunnels_active",
		Help:      "当前活跃的隧道监听器数（监听中的 TCP/UDP 端口）",
	})

	// ClientsOnline 当前在线的客户端数
	ClientsOnline = prometheus.NewGauge(prometheus.GaugeOpts{
		Namespace: namespace,
		Name:      "clients_online",
		Help:      "当前在线的客户端数",
	})

	// ClientsTotal 注册客户端总数
	ClientsTotal = prometheus.NewGauge(prometheus.GaugeOpts{
		Namespace: namespace,
		Name:      "clients_total",
		Help:      "注册客户端总数",
	})

	// SessionsTotal 累计 session 数（含已关闭）
	SessionsTotal = prometheus.NewCounter(prometheus.CounterOpts{
		Namespace: namespace,
		Name:      "sessions_total",
		Help:      "累计 session 数",
	})

	// TunnelsBytesSent 隧道累计发送字节数
	TunnelsBytesSent = prometheus.NewCounterVec(prometheus.CounterOpts{
		Namespace: namespace,
		Name:      "tunnel_bytes_sent_total",
		Help:      "隧道累计发送字节数",
	}, []string{"tunnel_type", "client_id"})

	// TunnelsBytesReceived 隧道累计接收字节数
	TunnelsBytesReceived = prometheus.NewCounterVec(prometheus.CounterOpts{
		Namespace: namespace,
		Name:      "tunnel_bytes_received_total",
		Help:      "隧道累计接收字节数",
	}, []string{"tunnel_type", "client_id"})
)

func init() {
	prometheus.MustRegister(
		ActiveSessions,
		ActiveTunnels,
		ClientsOnline,
		ClientsTotal,
		SessionsTotal,
		TunnelsBytesSent,
		TunnelsBytesReceived,
	)
}
