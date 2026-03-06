package resolver

// ResolverProvider provides access to all resolvers.
type ResolverProvider interface {
	User() UserResolver
	Device() DeviceResolver
}
