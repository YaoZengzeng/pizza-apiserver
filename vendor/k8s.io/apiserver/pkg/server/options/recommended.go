/*
Copyright 2016 The Kubernetes Authors.

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

package options

import (
	"github.com/spf13/pflag"

	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apiserver/pkg/admission"
	"k8s.io/apiserver/pkg/server"
	"k8s.io/apiserver/pkg/storage/storagebackend"
)

// RecommendedOptions contains the recommended options for running an API server.
// RecommendedOptions包含运行一个API sever推荐的配置
// If you add something to this list, it should be in a logical grouping.
// Each of them can be nil to leave the feature unconfigured on ApplyTo.
type RecommendedOptions struct {
	// 用于从Etcd读写的配置
	Etcd           *EtcdOptions
	// 和https相关的配置，例如ports，certificates等
	SecureServing  *SecureServingOptionsWithLoopback
	Authentication *DelegatingAuthenticationOptions
	Authorization  *DelegatingAuthorizationOptions
	// 设置auditing output stack，默认为false，但是可以配置输出到一个audit log file或者
	// 发送audit event到一个external backend
	Audit          *AuditOptions
	Features       *FeatureOptions
	// 包含到可以访问main API server的kubeconfig的路径，默认使用in-cluster config
	CoreAPI        *CoreAPIOptions

	// ExtraAdmissionInitializers is called once after all ApplyTo from the options above, to pass the returned
	// admission plugin initializers to Admission.ApplyTo.
	// ExtraAdmissionInitializers允许为admission添加更多的initializer，
	ExtraAdmissionInitializers func(c *server.RecommendedConfig) ([]admission.PluginInitializer, error)
	Admission                  *AdmissionOptions
	// ProcessInfo is used to identify events created by the server.
	// ProcessInfo用于识别server产生的event
	ProcessInfo *ProcessInfo
	// authentication和admission webhook的配置
	Webhook     *WebhookOptions
}

func NewRecommendedOptions(prefix string, codec runtime.Codec, processInfo *ProcessInfo) *RecommendedOptions {
	sso := NewSecureServingOptions()

	// We are composing recommended options for an aggregated api-server,
	// whose client is typically a proxy multiplexing many operations ---
	// notably including long-running ones --- into one HTTP/2 connection
	// into this server.  So allow many concurrent operations.
	sso.HTTP2MaxStreamsPerConnection = 1000

	return &RecommendedOptions{
		// 读写etcd的backend
		Etcd:                       NewEtcdOptions(storagebackend.NewDefaultConfig(prefix, codec)),
		// SecureServing配置了所有和https相关的事情
		SecureServing:              sso.WithLoopback(),
		Authentication:             NewDelegatingAuthenticationOptions(),
		Authorization:              NewDelegatingAuthorizationOptions(),
		// Audit设置auditing output stack，默认关闭，但是可以配置发送到audit log file或者发送audit events到一个external backend
		Audit:                      NewAuditOptions(),
		Features:                   NewFeatureOptions(),
		// 访问Main API Server的配置文件
		CoreAPI:                    NewCoreAPIOptions(),
		// ExtraAdmissionInitializers允许我们增加更多的initializers用于实现informers以及client到custom APIServer之间的管道
		ExtraAdmissionInitializers: func(c *server.RecommendedConfig) ([]admission.PluginInitializer, error) { return nil, nil },
		// Admission是一个mutating和validating admission的插件栈，会对每一个incoming API request进行处理
		Admission:                  NewAdmissionOptions(),
		// ProcessInfo维护了event对象创建的信息
		ProcessInfo:                processInfo,
		// Webhook配置webhooks的操作（例如，设置authentication以及admission webhook）
		// 对于运行在集群中的custom API Servermore配置是足够了的
		Webhook:                    NewWebhookOptions(),
	}
}

func (o *RecommendedOptions) AddFlags(fs *pflag.FlagSet) {
	o.Etcd.AddFlags(fs)
	o.SecureServing.AddFlags(fs)
	o.Authentication.AddFlags(fs)
	o.Authorization.AddFlags(fs)
	o.Audit.AddFlags(fs)
	o.Features.AddFlags(fs)
	o.CoreAPI.AddFlags(fs)
	o.Admission.AddFlags(fs)
}

// ApplyTo adds RecommendedOptions to the server configuration.
// ApplyTo增加RecommendedOptions到server configuration
// pluginInitializers can be empty, it is only need for additional initializers.
func (o *RecommendedOptions) ApplyTo(config *server.RecommendedConfig) error {
	// 将各种配置应用到默认配置中
	if err := o.Etcd.ApplyTo(&config.Config); err != nil {
		return err
	}
	if err := o.SecureServing.ApplyTo(&config.Config.SecureServing, &config.Config.LoopbackClientConfig); err != nil {
		return err
	}
	if err := o.Authentication.ApplyTo(&config.Config.Authentication, config.SecureServing, config.OpenAPIConfig); err != nil {
		return err
	}
	if err := o.Authorization.ApplyTo(&config.Config.Authorization); err != nil {
		return err
	}
	if err := o.Audit.ApplyTo(&config.Config, config.ClientConfig, config.SharedInformerFactory, o.ProcessInfo, o.Webhook); err != nil {
		return err
	}
	if err := o.Features.ApplyTo(&config.Config); err != nil {
		return err
	}
	if err := o.CoreAPI.ApplyTo(config); err != nil {
		return err
	}
	// 获取initializer，再用initializer来初始化admission plugins
	if initializers, err := o.ExtraAdmissionInitializers(config); err != nil {
		return err
	} else if err := o.Admission.ApplyTo(&config.Config, config.SharedInformerFactory, config.ClientConfig, initializers...); err != nil {
		return err
	}

	return nil
}

func (o *RecommendedOptions) Validate() []error {
	errors := []error{}
	errors = append(errors, o.Etcd.Validate()...)
	errors = append(errors, o.SecureServing.Validate()...)
	errors = append(errors, o.Authentication.Validate()...)
	errors = append(errors, o.Authorization.Validate()...)
	errors = append(errors, o.Audit.Validate()...)
	errors = append(errors, o.Features.Validate()...)
	errors = append(errors, o.CoreAPI.Validate()...)
	errors = append(errors, o.Admission.Validate()...)

	return errors
}
