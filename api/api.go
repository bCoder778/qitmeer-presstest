package api

import (
	"fmt"
	"git.diabin.com/BlockChain/qitmeer-presstest/config"
	"git.diabin.com/BlockChain/qitmeer-presstest/controller"
	"strconv"
)

type Api struct {
	rest       *RestApi
	controller *controller.Controller
	conf       *config.Config
}

func NewApi(conf *config.Config) (*Api, error) {
	controller, err := controller.NewController(conf)
	if err != nil {
		return nil, err
	}
	rest := NewRestApi(conf)
	return &Api{
		rest:       rest,
		controller: controller,
		conf:       conf,
	}, nil
}

func (a *Api) Run() error {
	a.addApi()
	return a.rest.Start()
}

func (a *Api) addApi() {
	a.rest.AuthRouteSet("api/v1/sendPress").Post(a.sendPress)
	a.rest.AuthRouteSet("api/v1/split").Post(a.split)
	a.rest.AuthRouteSet("api/v1/addAddress").Post(a.addAddress)
	a.rest.AuthRouteSet("api/v1/transferAll").Post(a.transferAll)
	a.rest.AuthRouteSet("api/v1/transfer").Post(a.transfer)
	a.rest.AuthRouteSet("api/v1/mergeUTXO").Post(a.mergeUTXO)
	a.rest.AuthRouteSet("api/v1/unspentCount").Get(a.unspentCount)

	a.rest.AuthRouteSet("api/v1/evm/shareMoney").Get(a.shareMoney)
	a.rest.AuthRouteSet("api/v1/evm/pressEVM").Get(a.pressEVM)
}

func (a *Api) sendPress(ct *Context) (interface{}, *Error) {
	address := ct.Form["to"]
	coin := ct.Form["coin"]
	strCount := ct.Form["count"]
	strTps := ct.Form["tps"]
	count, err := strconv.Atoi(strCount)
	tps, err := strconv.Atoi(strTps)
	if err != nil {
		return nil, &Error{
			Code:    ERROR_UNKNOWN,
			Message: err.Error(),
		}
	}
	re := a.controller.SendPress(address, coin, count, tps)
	return re, nil
}

func (a *Api) split(ct *Context) (interface{}, *Error) {
	coin := ct.Form["coin"]
	strCount := ct.Form["count"]
	count, err := strconv.Atoi(strCount)
	if err != nil {
		return nil, &Error{
			Code:    ERROR_UNKNOWN,
			Message: err.Error(),
		}
	}
	re := a.controller.Split(coin, count)
	return re, nil
}

func (a *Api) addAddress(ct *Context) (interface{}, *Error) {
	strCount := ct.Form["count"]
	count, err := strconv.Atoi(strCount)
	if err != nil {
		count = 1
	}
	rs := a.controller.AddAddress(count)
	return rs, nil
}

func (a *Api) transferAll(ct *Context) (interface{}, *Error) {
	strAmount := ct.Form["amount"]
	from := ct.Form["from"]
	key := ct.Form["key"]
	amount, err := strconv.ParseUint(strAmount, 10, 64)
	if err != nil {
		return nil, &Error{
			Code:    ERROR_UNKNOWN,
			Message: err.Error(),
		}
	}
	rs := a.controller.TransferAll(from, key, amount)
	return rs, nil
}

func (a *Api) transfer(ct *Context) (interface{}, *Error) {
	strAmount := ct.Form["amount"]
	from := ct.Form["from"]
	to := ct.Form["to"]
	key := ct.Form["key"]
	amount, err := strconv.ParseUint(strAmount, 10, 64)
	if err != nil {
		return nil, &Error{
			Code:    ERROR_UNKNOWN,
			Message: err.Error(),
		}
	}
	rs := a.controller.Transfer(from, to, key, amount)
	return rs, nil
}

func (a *Api) mergeUTXO(ct *Context) (interface{}, *Error) {
	from := ct.Form["from"]
	key := ct.Form["key"]

	rs := a.controller.MergerUTXO(from, key)
	return rs, nil
}

func (a *Api) unspentCount(ct *Context) (interface{}, *Error) {
	rs := a.controller.UnspentCount()
	return rs, nil
}

func (a *Api) shareMoney(ct *Context) (interface{}, *Error) {
	amount := ct.Query["amount"]
	if amount == "" {
		amount = "100"
	}
	rs := a.controller.ShareMoney(amount)
	return rs, nil
}

func (a *Api) pressEVM(ct *Context) (interface{}, *Error) {
	strCount := ct.Query["count"]
	strTps := ct.Query["tps"]
	count, err := strconv.Atoi(strCount)
	tps, err := strconv.Atoi(strTps)
	if err != nil {
		return nil, &Error{
			Code:    ERROR_UNKNOWN,
			Message: err.Error(),
		}
	}
	re := a.controller.PressEVM(count, tps)
	return re, nil
}

func (a *Api) Stop() {
	if a.controller == nil {
		fmt.Println(a.controller)
	}
	a.controller.Stop()
	a.rest.Stop()
}
