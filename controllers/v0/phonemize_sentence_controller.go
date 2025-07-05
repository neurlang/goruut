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
	AllControllers["/phonemize/sentence"] = &PhonemizeSentenceController{}
}

type PhonemizeSentenceController struct {
	uc usecases.IPhonemizeUsecase
}

func (c *PhonemizeSentenceController) BackendType() ControllerBackendType {
	return MainController
}

func (c *PhonemizeSentenceController) ServeHTTP(w http.ResponseWriter, request *http.Request) {

	if request.Method != "POST" {
		w.WriteHeader(500)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	decoder := json.NewDecoder(request.Body)
	var req requests.PhonemizeSentence
	err := decoder.Decode(&req)
	if err != nil {
		log.Error0(err)
		w.WriteHeader(500)
		return
	}
	res := c.uc.Sentence(req)

	w.WriteHeader(200)
	log.Error0(helpers.Write(w, log.Error1(helpers.SerializeJson(res))))
}

func (c *PhonemizeSentenceController) Init(di *DependencyInjection) {
	usecase := MustNeed(di, usecases.NewPhonemizeUsecase)
	c.uc = &usecase
	di.Add(c)
}
