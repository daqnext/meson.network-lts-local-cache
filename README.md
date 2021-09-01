# meson.network-lts-local-cache

```
import (
	localcache "github.com/daqnext/meson.network-lts-local-cache"
	"log"
)

type Person struct {
	Name string
	Age int
}

func main() {
	lc:=localcache.New(0)
	lc.SetCountLimit(10000) //if not set default is 100000

	//set
	lc.Set("foo","bar",300)
	lc.Set("a",1,300)
	lc.Set("b",Person{"Jack",18},300)
	lc.Set("b*",&Person{"Jack",18},300)
	lc.Set("c",true,100) //never expire

	//get
	log.Println("---get---")
	log.Println(lc.Get("foo"))
	log.Println(lc.Get("a"))
	log.Println(lc.Get("b"))
	log.Println(lc.Get("b*"))
	log.Println(lc.Get("c"))


	//set cover
	log.Println("---cover set---")
	log.Println(lc.Get("c"))
	lc.Set("c",false,60) //never expire
	log.Println(lc.Get("c"))

}

```