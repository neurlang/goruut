package v0

import (
	"encoding/json"
	. "github.com/neurlang/goruut/controllers"
	"github.com/neurlang/goruut/helpers"
	"github.com/neurlang/goruut/helpers/log"
	"github.com/neurlang/goruut/models/requests"
	"github.com/neurlang/goruut/usecases"
	"net/http"
)
import . "github.com/martinarisk/di/dependency_injection"

func init() {
	AllControllers["/explain/word"] = &ExplainWordController{}
}

type ExplainWordController struct {
	uc usecases.IPhonemizeUsecase
}

func (c *ExplainWordController) BackendType() ControllerBackendType {
	return MainController
}

func (c *ExplainWordController) ServeHTTP(w http.ResponseWriter, request *http.Request) {

	if request.Method != "POST" {
		w.WriteHeader(500)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	decoder := json.NewDecoder(request.Body)
	var req requests.ExplainWord
	err := decoder.Decode(&req)
	if err != nil {
		w.WriteHeader(500)
		return
	}
	res := c.uc.Word(req)

	w.WriteHeader(200)
	log.Error0(helpers.Write(w, log.Error1(helpers.SerializeJson(res))))
}

func (c *ExplainWordController) Init(di *DependencyInjection) {
	usecase := MustNeed(di, usecases.NewPhonemizeUsecase)
	c.uc = &usecase
	di.Add(c)
}
