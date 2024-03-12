package storage

import (
	"github.com/Masterminds/semver"

	"helm.sh/helm/v3/pkg/chart"

	"github.com/kyma-project/helm-broker/internal"
)

// Addon is an interface that describe storage layer operations for Addons
type Addon interface {
	Upsert(internal.Namespace, *internal.Addon) (replace bool, err error)
	Get(internal.Namespace, internal.AddonName, semver.Version) (*internal.Addon, error)
	GetByID(internal.Namespace, internal.AddonID) (*internal.Addon, error)
	Remove(internal.Namespace, internal.AddonName, semver.Version) error
	RemoveByID(internal.Namespace, internal.AddonID) error
	RemoveAll(internal.Namespace) error
	FindAll(internal.Namespace) ([]*internal.Addon, error)
}

// Chart is an interface that describe storage layer operations for Charts
type Chart interface {
	Upsert(internal.Namespace, *chart.Chart) (replace bool, err error)
	Get(internal.Namespace, internal.ChartName, semver.Version) (*chart.Chart, error)
	Remove(internal.Namespace, internal.ChartName, semver.Version) error
}

// Instance is an interface that describe storage layer operations for Instances
type Instance interface {
	Insert(*internal.Instance) error
	Upsert(*internal.Instance) (replace bool, err error)
	Get(internal.InstanceID) (*internal.Instance, error)
	GetAll() ([]*internal.Instance, error)
	Remove(internal.InstanceID) error
}

// InstanceOperation is an interface that describe storage layer operations for InstanceOperations
type InstanceOperation interface {
	// Insert is inserting object into storage.
	// Object is modified by setting CreatedAt.
	Insert(*internal.InstanceOperation) error
	Get(internal.InstanceID, internal.OperationID) (*internal.InstanceOperation, error)
	GetAll(internal.InstanceID) ([]*internal.InstanceOperation, error)
	UpdateState(internal.InstanceID, internal.OperationID, internal.OperationState) error
	UpdateStateDesc(internal.InstanceID, internal.OperationID, internal.OperationState, *string) error
	Remove(internal.InstanceID, internal.OperationID) error
}

// BindOperation is an interface that describe storage layer operations for BindOperations
type BindOperation interface {
	Insert(*internal.BindOperation) error
	Get(internal.InstanceID, internal.BindingID, internal.OperationID) (*internal.BindOperation, error)
	GetAll(internal.InstanceID) ([]*internal.BindOperation, error)
	UpdateState(internal.InstanceID, internal.BindingID, internal.OperationID, internal.OperationState) error
	UpdateStateDesc(internal.InstanceID, internal.BindingID, internal.OperationID, internal.OperationState, *string) error
	Remove(internal.InstanceID, internal.BindingID, internal.OperationID) error
}

// InstanceBindData is an interface that describe storage layer operations for InstanceBindData entities
type InstanceBindData interface {
	Insert(*internal.InstanceBindData) error
	Get(internal.InstanceID) (*internal.InstanceBindData, error)
	Remove(internal.InstanceID) error
}

// IsNotFoundError checks if given error is NotFound error
func IsNotFoundError(err error) bool {
	nfe, ok := err.(interface {
		NotFound() bool
	})
	return ok && nfe.NotFound()
}

// IsAlreadyExistsError checks if given error is AlreadyExist error
func IsAlreadyExistsError(err error) bool {
	aee, ok := err.(interface {
		AlreadyExists() bool
	})
	return ok && aee.AlreadyExists()
}

// IsActiveOperationInProgressError checks if given error is ActiveOperationInProgress error
func IsActiveOperationInProgressError(err error) bool {
	aee, ok := err.(interface {
		ActiveOperationInProgress() bool
	})
	return ok && aee.ActiveOperationInProgress()
}
