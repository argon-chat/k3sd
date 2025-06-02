package addons

type AddonConfigBuilder interface {
	BuildConfig(domain string, subs map[string]string) map[string]interface{}
}

type AddonConfigBuilderFunc func(domain string, subs map[string]string) map[string]interface{}

func (f AddonConfigBuilderFunc) BuildConfig(domain string, subs map[string]string) map[string]interface{} {
	return f(domain, subs)
}

var AddonConfigBuilderRegistry = map[string]AddonConfigBuilder{}

func RegisterAddonConfigBuilder(name string, builder AddonConfigBuilder) {
	AddonConfigBuilderRegistry[name] = builder
}
