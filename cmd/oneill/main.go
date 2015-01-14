package main

import (
	"github.com/rehabstudio/oneill/oneill"
	"github.com/rehabstudio/oneill/processors"
)

func init() {
	oneill.Initialise()
}

func main() {

	// initialise and populate the pipeline with processors
	pipeline := oneill.ProcessorPipeline{}
	// load all site definitions from disk
	pipeline.AddProcessor(processors.LoadSiteDefinitions)
	// validate site definitions, removing any that don't pass
	pipeline.AddProcessor(processors.ValidateDefinitionsPrePull)
	// pull all defined images to ensure we have the latest versions locally
	pipeline.AddProcessor(processors.PullImages)
	// validate site definitions, removing any that don't pass
	pipeline.AddProcessor(processors.ValidateDefinitionsPostPull)
	// remove any containers that don't match the definition
	pipeline.AddProcessor(processors.RemoveContainers)
	// start any defined containers that aren't already running
	pipeline.AddProcessor(processors.StartContainers)
	// build nginx templates and reload the nginx service
	pipeline.AddProcessor(processors.ConfigureNginx)

	// run the processing pipeline
	pipeline.RunPipeline()

}
