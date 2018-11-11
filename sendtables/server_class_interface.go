// DO NOT EDIT: Auto generated

package sendtables

// IServerClass is an auto-generated interface for ServerClass.
// ServerClass stores meta information about Entity types (e.g. palyers, teams etc.).
type IServerClass interface {
	// ID returns the server-class's ID.
	ID() int
	// Name returns the server-class's name.
	Name() string
	// DataTableID returns the data-table ID.
	DataTableID() int
	// DataTableName returns the data-table name.
	DataTableName() string
	// BaseClasses returns the base-classes of this server-class.
	BaseClasses() []*ServerClass
	// PropertyEntries returns the names of all property-entries on this server-class.
	PropertyEntries() []string
	// OnEntityCreated registers a function to be called when a new entity is created from this ServerClass.
	OnEntityCreated(handler EntityCreatedHandler)
	String() string
}
