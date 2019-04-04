package main

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"os"

	"github.com/cloudfoundry-community/go-cfclient"
	. "github.com/logrusorgru/aurora"
	"github.com/urfave/cli"
)

func main() {

	app := cli.NewApp()

	app.Name = "cf-tools"
	app.Version = "0.1"
	app.Usage = "a set of useful commands which can query a local cloud-controller db cache for increased usage speed."

	app.Commands = []cli.Command{
		{
			Name:  "sync",
			Usage: "sync the local cache against target foundation",
			Action: func(c *cli.Context) error {
				syncCache()
				return nil
			},
		},
		{
			Name:  "service",
			Usage: "commands to investigate service instances",
			Subcommands: []cli.Command{
				{
					Name:  "list",
					Usage: "list available service types",
					Action: func(c *cli.Context) error {
						showServiceList()
						return nil
					},
				},
				{
					Name:  "usage",
					Usage: "shows service instance usage of target service type",
					Action: func(c *cli.Context) error {
						showServiceTree(c.Args().First())
						return nil
					},
				},
				{
					Name:  "get-guid",
					Usage: "search for service instance guid by name",
					Action: func(c *cli.Context) error {
						findServiceGUIDByServiceInstanceName(c.Args().First())

						return nil
					},
				},
			},
		},
		{
			Name:    "binding",
			Aliases: []string{"b"},
			Usage:   "commands to investigate service bindings",
			Subcommands: []cli.Command{
				{
					Name:  "app",
					Usage: "find binding by app guid",
					Action: func(c *cli.Context) error {
						findBindingByApp(c.Args().First())

						return nil
					},
				},
				{
					Name:  "service",
					Usage: "find binding by service instance guid",
					Action: func(c *cli.Context) error {
						findBindingByService(c.Args().First())

						return nil
					},
				},
			},
		},
		{
			Name:    "app",
			Aliases: []string{"a"},
			Usage:   "commands to investigate apps",
			Subcommands: []cli.Command{
				{
					Name:  "get-guid",
					Usage: "search for app guid by name",
					Action: func(c *cli.Context) error {
						findAppGUIDByAppName(c.Args().First())

						return nil
					},
				},
				{
					Name:  "show",
					Usage: "shows app info based on app guid",
					Action: func(c *cli.Context) error {
						findAppByAppGUID(c.Args().First())

						return nil
					},
				},
				{
					Name:  "health-check",
					Usage: "shows global info regarding number of crashed apps",
					Action: func(c *cli.Context) error {
						checkAppHealth()

						return nil
					},
				},
			},
		},
	}

	err := app.Run(os.Args)
	if err != nil {
		log.Fatal(err)
	}
}

func checkAppHealth() {
	cache := Cache{}
	cache.loadCache()

	fmt.Println()
	fmt.Println("Checking app health for the foundation.")
	fmt.Println()

	appsstarted := 0
	appsstopped := 0
	appsunhealthy := 0
	appscrashed := 0

	unhealthyApps := []cfclient.AppSummary{}
	crashedApps := []cfclient.AppSummary{}

	for appsummarycounter := 0; appsummarycounter < len(cache.appSummaries); appsummarycounter++ {
		if cache.appSummaries[appsummarycounter].State == "STARTED" {
			appsstarted++
		} else if cache.appSummaries[appsummarycounter].State == "STOPPED" {
			appsstopped++
		}

		if cache.appSummaries[appsummarycounter].RunningInstances == 0 && cache.appSummaries[appsummarycounter].State != "STOPPED" {
			crashedApps = append(crashedApps, cache.appSummaries[appsummarycounter])
			appscrashed++
		} else if cache.appSummaries[appsummarycounter].RunningInstances < cache.appSummaries[appsummarycounter].Instances && cache.appSummaries[appsummarycounter].State != "STOPPED" {
			appsunhealthy++
			unhealthyApps = append(unhealthyApps, cache.appSummaries[appsummarycounter])
		}
	}

	if appscrashed > 0 {
		fmt.Println("------------------------")
		fmt.Println()
		fmt.Println(Bold(Red("Crashed Apps")))
		fmt.Println()

		for counter := 0; counter < len(crashedApps); counter++ {
			for spacecounter := 0; spacecounter < len(cache.spaces); spacecounter++ {
				if crashedApps[counter].SpaceGuid == cache.spaces[spacecounter].Guid {
					for orgcounter := 0; orgcounter < len(cache.orgs); orgcounter++ {
						if cache.spaces[spacecounter].OrganizationGuid == cache.orgs[orgcounter].Guid {
							fmt.Println("Org: ", cache.orgs[orgcounter].Name)
							fmt.Println("Space: ", cache.spaces[spacecounter].Name)
							fmt.Println("App Name: ", crashedApps[counter].Name)
							fmt.Println("App Guid: ", crashedApps[counter].Guid)
							fmt.Println("App State: ", crashedApps[counter].State)
							fmt.Println("Instances: ", crashedApps[counter].Instances)
							fmt.Println("Running Instances: ", crashedApps[counter].RunningInstances)
							fmt.Println()
						}
					}
				}
			}
		}
	}

	if appsunhealthy > 0 {
		fmt.Println("------------------------")
		fmt.Println()
		fmt.Println(Bold(Cyan("Unhealthy Apps")))
		fmt.Println()

		for counter := 0; counter < len(unhealthyApps); counter++ {
			for spacecounter := 0; spacecounter < len(cache.spaces); spacecounter++ {
				if unhealthyApps[counter].SpaceGuid == cache.spaces[spacecounter].Guid {
					for orgcounter := 0; orgcounter < len(cache.orgs); orgcounter++ {
						if cache.spaces[spacecounter].OrganizationGuid == cache.orgs[orgcounter].Guid {
							fmt.Println("Org: ", cache.orgs[orgcounter].Name)
							fmt.Println("Space: ", cache.spaces[spacecounter].Name)
							fmt.Println("App Name: ", unhealthyApps[counter].Name)
							fmt.Println("App Guid: ", unhealthyApps[counter].Guid)
							fmt.Println("App State: ", unhealthyApps[counter].State)
							fmt.Println("Instances: ", unhealthyApps[counter].Instances)
							fmt.Println("Running Instances: ", unhealthyApps[counter].RunningInstances)
							fmt.Println()
						}
					}
				}
			}
		}
	}

	fmt.Println("------------------------")
	fmt.Println()

	fmt.Println("Total number of apps: ", len(cache.apps))
	fmt.Println("Total number of running: ", appsstarted)
	fmt.Println("Total number of stopped: ", appsstopped)
	fmt.Println("Total number of unhealthy apps: ", appsunhealthy)
	fmt.Println("Total number of crashed apps: ", appscrashed)
}

// TO DO
func findAppByAppGUID(guid string) {
	cache := Cache{}
	cache.loadCache()
	fmt.Println("Searching for app by app guid: ", guid)
}

func findAppGUIDByAppName(name string) {
	cache := Cache{}
	cache.loadCache()

	fmt.Println()
	fmt.Println("Searching for app guid by app name: ", name)
	fmt.Println()

	for appcounter := 0; appcounter < len(cache.apps); appcounter++ {
		if cache.apps[appcounter].Name == name {
			for spacecounter := 0; spacecounter < len(cache.spaces); spacecounter++ {
				if cache.apps[appcounter].SpaceGuid == cache.spaces[spacecounter].Guid {
					for orgcounter := 0; orgcounter < len(cache.orgs); orgcounter++ {
						if cache.spaces[spacecounter].OrganizationGuid == cache.orgs[orgcounter].Guid {
							fmt.Println("Org: ", cache.orgs[orgcounter].Name)
							fmt.Println("Space: ", cache.spaces[spacecounter].Name)
							fmt.Println("App Name: ", cache.apps[appcounter].Name)
							fmt.Println("App Guid: ", cache.apps[appcounter].Guid)
							fmt.Println("App State: ", cache.apps[appcounter].State)
							fmt.Println()
						}
					}
				}
			}
		}
	}
}

func findBindingByService(guid string) {
	cache := Cache{}
	cache.loadCache()

	fmt.Println()
	fmt.Println("Searching for bindings by service instance guid: ", guid)
	fmt.Println()

	for servicebindingcounter := 0; servicebindingcounter < len(cache.serviceBindings); servicebindingcounter++ {
		if cache.serviceBindings[servicebindingcounter].ServiceInstanceGuid == guid {
			for appcounter := 0; appcounter < len(cache.apps); appcounter++ {
				if cache.serviceBindings[servicebindingcounter].AppGuid == cache.apps[appcounter].Guid {
					for spacecounter := 0; spacecounter < len(cache.spaces); spacecounter++ {
						if cache.apps[appcounter].SpaceGuid == cache.spaces[spacecounter].Guid {
							for orgcounter := 0; orgcounter < len(cache.orgs); orgcounter++ {
								if cache.spaces[spacecounter].OrganizationGuid == cache.orgs[orgcounter].Guid {
									fmt.Println("Org: ", cache.orgs[orgcounter].Name)
									fmt.Println("Space: ", cache.spaces[spacecounter].Name)
									fmt.Println("App Name: ", cache.apps[appcounter].Name)
									fmt.Println("App Guid: ", cache.apps[appcounter].Guid)
									fmt.Println()
								}
							}
						}
					}
				}
			}
		}
	}
}

func findBindingByApp(guid string) {
	cache := Cache{}
	cache.loadCache()

	fmt.Println()
	fmt.Println("Searching for bindings by app guid: ", guid)
	fmt.Println()

	for servicebindingcounter := 0; servicebindingcounter < len(cache.serviceBindings); servicebindingcounter++ {
		if cache.serviceBindings[servicebindingcounter].AppGuid == guid {
			for serviceinstancecounter := 0; serviceinstancecounter < len(cache.serviceInstances); serviceinstancecounter++ {
				if cache.serviceBindings[servicebindingcounter].ServiceInstanceGuid == cache.serviceInstances[serviceinstancecounter].Guid {
					for spacecounter := 0; spacecounter < len(cache.spaces); spacecounter++ {
						if cache.serviceInstances[serviceinstancecounter].SpaceGuid == cache.spaces[spacecounter].Guid {
							for orgcounter := 0; orgcounter < len(cache.orgs); orgcounter++ {
								if cache.spaces[spacecounter].OrganizationGuid == cache.orgs[orgcounter].Guid {
									fmt.Println("Org: ", cache.orgs[orgcounter].Name)
									fmt.Println("Space: ", cache.spaces[spacecounter].Name)
									fmt.Println("Service Name: ", cache.serviceInstances[serviceinstancecounter].Name)
									fmt.Println("Service Guid: ", cache.serviceInstances[serviceinstancecounter].Guid)
									fmt.Println()
								}
							}
						}
					}
				}
			}
		}
	}
}

func findServiceGUIDByServiceInstanceName(name string) {
	cache := Cache{}
	cache.loadCache()

	fmt.Println()
	fmt.Println("Searching for service guid by service instance name: ", name)
	fmt.Println()

	for serviceinstancecounter := 0; serviceinstancecounter < len(cache.serviceInstances); serviceinstancecounter++ {
		if cache.serviceInstances[serviceinstancecounter].Name == name {
			for spacecounter := 0; spacecounter < len(cache.spaces); spacecounter++ {
				if cache.serviceInstances[serviceinstancecounter].SpaceGuid == cache.spaces[spacecounter].Guid {
					for orgcounter := 0; orgcounter < len(cache.orgs); orgcounter++ {
						if cache.spaces[spacecounter].OrganizationGuid == cache.orgs[orgcounter].Guid {
							fmt.Println("Org: ", cache.orgs[orgcounter].Name)
							fmt.Println("Space: ", cache.spaces[spacecounter].Name)
							fmt.Println("Service Name: ", cache.serviceInstances[serviceinstancecounter].Name)
							fmt.Println("Service Guid: ", cache.serviceInstances[serviceinstancecounter].Guid)
							fmt.Println()
						}
					}
				}
			}
		}
	}
}

func showServiceList() {
	cache := Cache{}
	cache.loadCache()

	// List service options
	fmt.Println()
	fmt.Println("Services available:")
	fmt.Println()
	for i := 0; i < len(cache.services); i++ {
		fmt.Println(Bold(cache.services[i].Label), " \n ", cache.services[i].Description)
	}
	fmt.Println()
}

func showServiceTree(search string) {
	cache := Cache{}
	cache.loadCache()

	fmt.Println()
	fmt.Println("You've entered:", search)
	fmt.Println()

	serviceGUID := ""

	for i := 0; i < len(cache.services); i++ {
		if cache.services[i].Label == search {
			serviceGUID = cache.services[i].Guid
		}
	}

	if serviceGUID == "" {
		fmt.Println("Could not find a service guid with your label. Please try again.")
		os.Exit(-1)
	}

	var matchingInstances []cfclient.ServiceInstance
	for i := 0; i < len(cache.serviceInstances); i++ {
		if cache.serviceInstances[i].ServiceGuid == serviceGUID {
			matchingInstances = append(matchingInstances, cache.serviceInstances[i])
		}
	}

	sortedList := []cfclient.ServiceInstance{}
	lastOrg := ""
	lastSpace := ""
	for o := 0; o < len(cache.orgs); o++ {
		for s := 0; s < len(cache.spaces); s++ {
			if cache.spaces[s].OrganizationGuid == cache.orgs[o].Guid {
				//match
				for i := 0; i < len(matchingInstances); i++ {
					if matchingInstances[i].SpaceGuid == cache.spaces[s].Guid {
						//match
						if lastOrg == cache.orgs[o].Guid && lastSpace == cache.spaces[s].Guid {
							sortedList = append(sortedList, matchingInstances[i])
							lastOrg = cache.orgs[o].Guid
							lastSpace = cache.spaces[s].Guid
						} else if lastOrg == cache.orgs[o].Guid {
							// new space, same org
							sortedList = append(sortedList, matchingInstances[i])
							lastOrg = cache.orgs[o].Guid
							lastSpace = cache.spaces[s].Guid
						} else {
							sortedList = append(sortedList, matchingInstances[i])
							lastOrg = cache.orgs[o].Guid
							lastSpace = cache.spaces[s].Guid
						}
					}
				}
			}
		}
	}

	lastOrg = ""
	lastSpace = ""
	for serviceinstancecounter := 0; serviceinstancecounter < len(sortedList); serviceinstancecounter++ {
		for spacecounter := 0; spacecounter < len(cache.spaces); spacecounter++ {
			if sortedList[serviceinstancecounter].SpaceGuid == cache.spaces[spacecounter].Guid {
				for orgcounter := 0; orgcounter < len(cache.orgs); orgcounter++ {
					if cache.spaces[spacecounter].OrganizationGuid == cache.orgs[orgcounter].Guid {
						if lastOrg == cache.orgs[orgcounter].Guid && lastSpace == cache.spaces[spacecounter].Guid {
							if serviceinstancecounter == (len(sortedList) - 1) {
								// final app of tree
								fmt.Println("+   └──", sortedList[serviceinstancecounter].Name)
								lastOrg = cache.orgs[orgcounter].Guid
								lastSpace = cache.spaces[spacecounter].Guid
							} else if sortedList[serviceinstancecounter].SpaceGuid == sortedList[serviceinstancecounter+1].SpaceGuid {
								// middle app of space
								fmt.Println("│   ├──", sortedList[serviceinstancecounter].Name)
								lastOrg = cache.orgs[orgcounter].Guid
								lastSpace = cache.spaces[spacecounter].Guid
							} else {
								// last app of space
								fmt.Println("│   └──", sortedList[serviceinstancecounter].Name)
								lastOrg = cache.orgs[orgcounter].Guid
								lastSpace = cache.spaces[spacecounter].Guid
							}
						} else if lastOrg == cache.orgs[orgcounter].Guid {
							if serviceinstancecounter == (len(sortedList) - 1) {
								// final app of tree, space
								fmt.Println("├──", Green(cache.spaces[spacecounter].Name))
								fmt.Println("+   └──", sortedList[serviceinstancecounter].Name)
								lastOrg = cache.orgs[orgcounter].Guid
								lastSpace = cache.spaces[spacecounter].Guid
							} else if sortedList[serviceinstancecounter].SpaceGuid != sortedList[serviceinstancecounter+1].SpaceGuid {
								// only app of space
								fmt.Println("├──", Green(cache.spaces[spacecounter].Name))
								fmt.Println("│   └──", sortedList[serviceinstancecounter].Name)
								lastOrg = cache.orgs[orgcounter].Guid
								lastSpace = cache.spaces[spacecounter].Guid
							} else {
								// first app of space
								fmt.Println("├──", Green(cache.spaces[spacecounter].Name))
								fmt.Println("│   ├──", sortedList[serviceinstancecounter].Name)
								lastOrg = cache.orgs[orgcounter].Guid
								lastSpace = cache.spaces[spacecounter].Guid
							}
						} else {
							if serviceinstancecounter == (len(sortedList) - 1) {
								// final app of tree, org and space
								fmt.Println(".", Bold(Cyan(cache.orgs[orgcounter].Name)))
								fmt.Println("├──", Green(cache.spaces[spacecounter].Name))
								fmt.Println("+   └──", sortedList[serviceinstancecounter].Name)
								lastOrg = cache.orgs[orgcounter].Guid
								lastSpace = cache.spaces[spacecounter].Guid
							} else if sortedList[serviceinstancecounter].SpaceGuid == sortedList[serviceinstancecounter+1].SpaceGuid {
								// first app of space
								fmt.Println(".", Bold(Cyan(cache.orgs[orgcounter].Name)))
								fmt.Println("├──", Green(cache.spaces[spacecounter].Name))
								fmt.Println("│   ├──", sortedList[serviceinstancecounter].Name)
								lastOrg = cache.orgs[orgcounter].Guid
								lastSpace = cache.spaces[spacecounter].Guid
							} else {
								// only app of space
								fmt.Println(".", Bold(Cyan(cache.orgs[orgcounter].Name)))
								fmt.Println("├──", Green(cache.spaces[spacecounter].Name))
								fmt.Println("│   └──", sortedList[serviceinstancecounter].Name)
								lastOrg = cache.orgs[orgcounter].Guid
								lastSpace = cache.spaces[spacecounter].Guid
							}
						}
					}
				}
			}
		}
	}

}

type Cache struct {
	orgs             []cfclient.Org
	spaces           []cfclient.Space
	apps             []cfclient.App
	appSummaries     []cfclient.AppSummary
	services         []cfclient.Service
	servicePlans     []cfclient.ServicePlan
	serviceInstances []cfclient.ServiceInstance
	serviceBindings  []cfclient.ServiceBinding
}

func (cache *Cache) loadCache() {
	//Import orgs to memory
	orgsFile, err := os.Open(os.Getenv("HOME") + "/.cfcache/orgs.json")

	if os.IsNotExist(err) {
		fmt.Println("orgs.json does not exist in the cache. Please run 'cf-tools sync'")
		orgsFile, err = os.Create(os.Getenv("HOME") + "/.cfcache/orgs.json")
	}
	defer orgsFile.Close()
	byteValue, _ := ioutil.ReadAll(orgsFile)
	//var orgs []cfclient.Org
	json.Unmarshal(byteValue, &cache.orgs)

	//Import spaces to memory
	spacesFile, err := os.Open(os.Getenv("HOME") + "/.cfcache/spaces.json")

	if os.IsNotExist(err) {
		fmt.Println("spaces.json does not exist in the cache. Please run 'cf-tools sync'")
		spacesFile, err = os.Create(os.Getenv("HOME") + "/.cfcache/spaces.json")
	}
	defer spacesFile.Close()
	byteValue, _ = ioutil.ReadAll(spacesFile)
	json.Unmarshal(byteValue, &cache.spaces)

	//Import apps to memory
	appsFile, err := os.Open(os.Getenv("HOME") + "/.cfcache/apps.json")

	if os.IsNotExist(err) {
		fmt.Println("apps.json does not exist in the cache. Please run 'cf-tools sync'")
		appsFile, err = os.Create(os.Getenv("HOME") + "/.cfcache/apps.json")
	}
	defer appsFile.Close()
	byteValue, _ = ioutil.ReadAll(appsFile)
	json.Unmarshal(byteValue, &cache.apps)

	//Import appSummaries to memory
	appSummariesFile, err := os.Open(os.Getenv("HOME") + "/.cfcache/appSummaries.json")

	if os.IsNotExist(err) {
		fmt.Println("appSummaries.json does not exist in the cache. Please run 'cf-tools sync'")
		appSummariesFile, err = os.Create(os.Getenv("HOME") + "/.cfcache/appSummaries.json")
	}
	defer appSummariesFile.Close()
	byteValue, _ = ioutil.ReadAll(appSummariesFile)
	json.Unmarshal(byteValue, &cache.appSummaries)

	//Import services to memory
	servicesFile, err := os.Open(os.Getenv("HOME") + "/.cfcache/services.json")

	if os.IsNotExist(err) {
		fmt.Println("services.json does not exist in the cache. Please run 'cf-tools sync'")
		servicesFile, err = os.Create(os.Getenv("HOME") + "/.cfcache/services.json")
	}
	defer servicesFile.Close()
	byteValue, _ = ioutil.ReadAll(servicesFile)
	json.Unmarshal(byteValue, &cache.services)

	//Import service plans to memory
	servicePlansFile, err := os.Open(os.Getenv("HOME") + "/.cfcache/servicePlans.json")

	if os.IsNotExist(err) {
		fmt.Println("servicePlans.json does not exist in the cache. Please run 'cf-tools sync'")
		servicePlansFile, err = os.Create(os.Getenv("HOME") + "/.cfcache/servicePlans.json")
	}
	defer servicePlansFile.Close()
	byteValue, _ = ioutil.ReadAll(servicePlansFile)
	json.Unmarshal(byteValue, &cache.servicePlans)

	//Import service instances to memory
	serviceInstancesFile, err := os.Open(os.Getenv("HOME") + "/.cfcache/serviceInstances.json")

	if os.IsNotExist(err) {
		fmt.Println("serviceInstances.json does not exist in the cache. Please run 'cf-tools sync'")
		serviceInstancesFile, err = os.Create(os.Getenv("HOME") + "/.cfcache/serviceInstances.json")
	}
	defer serviceInstancesFile.Close()
	byteValue, _ = ioutil.ReadAll(serviceInstancesFile)
	json.Unmarshal(byteValue, &cache.serviceInstances)

	//Import service bindings to memory
	serviceBindingsFile, err := os.Open(os.Getenv("HOME") + "/.cfcache/serviceBindings.json")

	if os.IsNotExist(err) {
		fmt.Println("serviceBindings.json does not exist in the cache. Please run 'cf-tools sync'")
		serviceInstancesFile, err = os.Create(os.Getenv("HOME") + "/.cfcache/serviceBindings.json")
	}
	defer serviceBindingsFile.Close()
	byteValue, _ = ioutil.ReadAll(serviceBindingsFile)
	json.Unmarshal(byteValue, &cache.serviceBindings)
}

func syncCache() {
	// Grab environment variables to form CF API Connection
	if os.Getenv("CF_API_ADDRESS") == "" || os.Getenv("CF_USERNAME") == "" || os.Getenv("CF_PASSWORD") == "" || os.Getenv("HOME") == "" {
		fmt.Printf("Please define env variables: CF_API_ADDRESS, CF_USERNAME, CF_PASSWORD, HOME")
		os.Exit(-1)
	}

	fmt.Println("Env variables look ok")

	c := &cfclient.Config{
		ApiAddress:        os.Getenv("CF_API_ADDRESS"),
		Username:          os.Getenv("CF_USERNAME"),
		Password:          os.Getenv("CF_PASSWORD"),
		SkipSslValidation: true,
	}

	fmt.Println("Creating cf client")

	client, _ := cfclient.NewClient(c)

	fmt.Println("Grabbing orgs from api")
	orgs, _ := client.ListOrgs()

	fmt.Println("Grabbing spaces from api")
	spaces, _ := client.ListSpaces()

	fmt.Println("Grabbing apps from api")
	apps, _ := client.ListApps()

	fmt.Println("Grabbing appSummaries from api")
	appSummaries := []cfclient.AppSummary{}

	for appcounter := 0; appcounter < len(apps); appcounter++ {
		toAdd, _ := apps[appcounter].Summary()
		appSummaries = append(appSummaries, toAdd)
	}

	fmt.Println("Grabbing services from api")
	services, _ := client.ListServices()

	fmt.Println("Grabbing servicePlans from api")
	servicePlans, _ := client.ListServicePlans()

	fmt.Println("Grabbing serviceInstances from api")
	serviceInstances, _ := client.ListServiceInstances()

	fmt.Println("Grabbing serviceBindings from api")
	serviceBindings, _ := client.ListServiceBindings()

	fmt.Println("Creating cache directory if it doesn't already exist")
	os.Mkdir(os.Getenv("HOME")+"/.cfcache", 0755)

	//Orgs Cache
	fmt.Println("Opening orgs.json")

	err := os.Remove(os.Getenv("HOME") + "/.cfcache/orgs.json")
	if err != nil {
		if os.IsNotExist(err) {
			fmt.Println("orgs.json does not exist and will be created.")
		} else {
			log.Fatal(err)
		}
	}

	orgcache, err := os.Create(os.Getenv("HOME") + "/.cfcache/orgs.json")

	defer orgcache.Close()

	fmt.Println("Writing orgs to file")
	towrite, err := json.Marshal(orgs)
	if err != nil {
		fmt.Println(err)
		return
	}
	orgcache.Write(towrite)

	//Spaces Cache
	fmt.Println("Opening spaces.json")

	err = os.Remove(os.Getenv("HOME") + "/.cfcache/spaces.json")
	if err != nil {
		if os.IsNotExist(err) {
			fmt.Println("spaces.json does not exist and will be created.")
		} else {
			log.Fatal(err)
		}
	}

	spacecache, err := os.Create(os.Getenv("HOME") + "/.cfcache/spaces.json")

	defer spacecache.Close()

	fmt.Println("Writing spaces to file")
	towrite, err = json.Marshal(spaces)
	if err != nil {
		fmt.Println(err)
		return
	}
	spacecache.Write(towrite)

	//App Cache
	fmt.Println("Opening apps.json")

	err = os.Remove(os.Getenv("HOME") + "/.cfcache/apps.json")
	if err != nil {
		if os.IsNotExist(err) {
			fmt.Println("apps.json does not exist and will be created.")
		} else {
			log.Fatal(err)
		}
	}

	appcache, err := os.Create(os.Getenv("HOME") + "/.cfcache/apps.json")

	defer appcache.Close()

	fmt.Println("Writing apps to file")
	towrite, err = json.Marshal(apps)
	if err != nil {
		fmt.Println(err)
		return
	}
	appcache.Write(towrite)

	//Services Cache
	fmt.Println("Opening services.json")

	err = os.Remove(os.Getenv("HOME") + "/.cfcache/services.json")
	if err != nil {
		if os.IsNotExist(err) {
			fmt.Println("services.json does not exist and will be created.")
		} else {
			log.Fatal(err)
		}
	}

	servicescache, err := os.Create(os.Getenv("HOME") + "/.cfcache/services.json")

	defer servicescache.Close()

	fmt.Println("Writing services to file")
	towrite, err = json.Marshal(services)
	if err != nil {
		fmt.Println(err)
		return
	}
	servicescache.Write(towrite)

	//servicePlans Cache
	fmt.Println("Opening servicePlans.json")

	err = os.Remove(os.Getenv("HOME") + "/.cfcache/servicePlans.json")
	if err != nil {
		if os.IsNotExist(err) {
			fmt.Println("servicePlans.json does not exist and will be created.")
		} else {
			log.Fatal(err)
		}
	}

	servicePlansCache, err := os.Create(os.Getenv("HOME") + "/.cfcache/servicePlans.json")

	defer servicePlansCache.Close()

	fmt.Println("Writing service plans to file")
	towrite, err = json.Marshal(servicePlans)
	if err != nil {
		fmt.Println(err)
		return
	}
	servicePlansCache.Write(towrite)

	//serviceInstances Cache
	fmt.Println("Opening serviceInstances.json")

	err = os.Remove(os.Getenv("HOME") + "/.cfcache/serviceInstances.json")
	if err != nil {
		if os.IsNotExist(err) {
			fmt.Println("serviceInstances.json does not exist and will be created.")
		} else {
			log.Fatal(err)
		}
	}

	serviceInstancesCache, err := os.Create(os.Getenv("HOME") + "/.cfcache/serviceInstances.json")

	defer serviceInstancesCache.Close()

	towrite, err = json.Marshal(serviceInstances)
	fmt.Println("Writing service instances to file")
	if err != nil {
		fmt.Println(err)
		return
	}
	serviceInstancesCache.Write(towrite)

	//ServiceBindings Cache
	fmt.Println("Opening serviceBindings.json")

	err = os.Remove(os.Getenv("HOME") + "/.cfcache/serviceBindings.json")
	if err != nil {
		if os.IsNotExist(err) {
			fmt.Println("serviceBindings.json does not exist and will be created.")
		} else {
			log.Fatal(err)
		}
	}

	serviceBindingsCache, err := os.Create(os.Getenv("HOME") + "/.cfcache/serviceBindings.json")

	defer serviceBindingsCache.Close()

	fmt.Println("Writing serviceBindings to file")
	towrite, err = json.Marshal(serviceBindings)
	if err != nil {
		fmt.Println(err)
		return
	}
	serviceBindingsCache.Write(towrite)

	//appSummaries Cache
	fmt.Println("Opening appSummaries.json")

	err = os.Remove(os.Getenv("HOME") + "/.cfcache/appSummaries.json")
	if err != nil {
		if os.IsNotExist(err) {
			fmt.Println("appSummaries.json does not exist and will be created.")
		} else {
			log.Fatal(err)
		}
	}

	appSummariesCache, err := os.Create(os.Getenv("HOME") + "/.cfcache/appSummaries.json")

	defer appSummariesCache.Close()

	fmt.Println("Writing appSummaries to file")
	towrite, err = json.Marshal(appSummaries)
	if err != nil {
		fmt.Println(err)
		return
	}
	appSummariesCache.Write(towrite)
}
