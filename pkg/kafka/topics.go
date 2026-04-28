package kafka

import pkgkafka "gitlab.com/lifegoeson-libs/pkg-gokit/kafka"

const (
	TopicProductCreated = "product.created"
	TopicProductUpdated = "product.updated"
	TopicProductDeleted = "product.deleted"
)

// EnvTopicResolver prefixes Kafka topics with an environment string.
type EnvTopicResolver struct {
	env    string
	topics []string
}

// NewEnvTopicResolver creates a resolver that prefixes all topics with "<env>.".
func NewEnvTopicResolver(env string, topics ...string) pkgkafka.TopicResolver {
	return &EnvTopicResolver{env: env, topics: topics}
}

func (r *EnvTopicResolver) ResolveTopic(baseTopic string) string {
	if r.env == "" {
		return baseTopic
	}
	return r.env + "." + baseTopic
}

func (r *EnvTopicResolver) GetAllTopics() []string {
	resolved := make([]string, len(r.topics))
	for i, t := range r.topics {
		resolved[i] = r.ResolveTopic(t)
	}
	return resolved
}
