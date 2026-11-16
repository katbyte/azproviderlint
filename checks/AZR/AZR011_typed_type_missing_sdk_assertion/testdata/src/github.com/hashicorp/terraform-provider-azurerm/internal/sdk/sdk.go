// Package sdk is a minimal stub of the provider's typed SDK: the interfaces it registers, cut down to enough methods to tell them apart.
package sdk

type ResourceFunc struct{}

type ResourceMetadata struct{}

type resourceBase interface {
	Arguments() map[string]string
	Attributes() map[string]string
	ModelObject() any
	ResourceType() string
}

type DataSource interface {
	resourceBase
	Read() ResourceFunc
}

type Resource interface {
	resourceBase
	Create() ResourceFunc
	Read() ResourceFunc
	Delete() ResourceFunc
	IDValidationFunc() func(string) error
}

type ResourceWithUpdate interface {
	Resource
	Update() ResourceFunc
}

type ResourceWithCustomizeDiff interface {
	Resource
	CustomizeDiff() ResourceFunc
}

type FrameworkWrappedResource interface {
	ModelObject() any
	ResourceType() string
	Create(metadata ResourceMetadata)
	Read(metadata ResourceMetadata)
	Delete(metadata ResourceMetadata)
}

type FrameworkWrappedResourceWithUpdate interface {
	FrameworkWrappedResource
	Update(metadata ResourceMetadata)
}

type FrameworkWrappedDataSource interface {
	ModelObject() any
	ResourceType() string
	Read(metadata ResourceMetadata)
}

type FrameworkListWrappedResource interface {
	List(metadata ResourceMetadata)
}

type EphemeralResource interface {
	Open()
}

type Action interface {
	Invoke()
}
