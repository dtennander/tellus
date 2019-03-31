package tellus

// Configuration that should be present in the file `.tellus` in any repository that should be served by Tellus.
type Configuration struct {
	// The directory that is the terraform root.
	// It is in this directory that Tellus will run all terraform commands.
	TerraformDirectory string `yaml:"tfDirectory"`
	// The branch on which the terraform should be applied.
	Branch string `yaml:"branch"`
}

func NewDefaultConfig() *Configuration {
	return &Configuration{
		TerraformDirectory: "",
		Branch: "master",
	}
}
