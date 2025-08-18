package pkg

type Generator struct {
	UUID        string
	Kind        string
	Name        string
	Description string
	Repository  Repository
	Params      Map
	Values      Map
}

type Repository struct {
	Kind   string
	URL    string
	Branch string
	Tag    string
	Path   string
}
