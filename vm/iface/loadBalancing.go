package iface

type LoadBalancingRegisterSelfConfig struct {
	Host  string
	Port  string
	EnvId string
}

type LoadBalancingRegisterBalancerConfig struct {
	Host string
	Port string
}

type LoadBalancingRegisterConfig struct {
	Self         LoadBalancingRegisterSelfConfig
	LoadBalancer LoadBalancingRegisterBalancerConfig
}
