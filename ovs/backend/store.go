package backend

type Store interface {
	Reserve(id, ovsIfaceName string) (bool, error)
	ReleaseByID(id string) (string, error)
}
