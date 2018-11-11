// DO NOT EDIT: Auto generated

package sendtables

// IServerClasses is an auto-generated interface for ServerClasses.
// ServerClasses is a searchable list of ServerClasses.
type IServerClasses interface {
	// FindByName finds and returns a server-class by it's name.
	//
	// Returns nil if the server-class wasn't found.
	//
	// Panics if more than one server-class with the same name was found.
	FindByName(name string) IServerClass
}
