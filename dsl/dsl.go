// From https://github.com/Tencent/AI-Infra-Guard
package dsl

// Config 定义了进行指纹匹配时需要的配置信息
type Config struct {
	Status int
	Body   string
	Header string
	Icon   int32
}
