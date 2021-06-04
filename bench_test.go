package jsexpr_test

import (
	"testing"
	"time"

	"github.com/byte-power/jsexpr"
	"github.com/byte-power/jsexpr/vm"
	"github.com/stretchr/testify/assert"
)

func Benchmark_simpleExpr(b *testing.B) {

	program, err := jsexpr.Compile(`1+2>2*0.5`)
	if err != nil {
		b.Fatal(err)
	}

	var out interface{}

	b.ResetTimer()
	for n := 0; n < b.N; n++ {
		out, err = vm.Run(program, nil)
	}
	b.StopTimer()

	if err != nil {
		b.Fatal(err)
	}
	if !out.(bool) {
		b.Fail()
	}
}

func Benchmark_expr(b *testing.B) {
	params := make(map[string]interface{})
	params["Origin"] = "MOW"
	params["Country"] = "RU"
	params["Adults"] = 1
	params["Value"] = 100

	program, err := jsexpr.Compile(`(Origin == "MOW" || Country == "RU") && (Value >= 100 || Adults == 1)`, jsexpr.TypeCheck(params))
	if err != nil {
		b.Fatal(err)
	}

	var out interface{}

	b.ResetTimer()
	for n := 0; n < b.N; n++ {
		out, err = vm.Run(program, params)
	}
	b.StopTimer()

	if err != nil {
		b.Fatal(err)
	}
	if !out.(bool) {
		b.Fail()
	}
}

func Benchmark_expr_reuseVm(b *testing.B) {
	params := make(map[string]interface{})
	params["Origin"] = "MOW"
	params["Country"] = "RU"
	params["Adults"] = 1
	params["Value"] = 100

	program, err := jsexpr.Compile(`(Origin == "MOW" || Country == "RU") && (Value >= 100 || Adults == 1)`, jsexpr.TypeCheck(params))
	if err != nil {
		b.Fatal(err)
	}

	var out interface{}
	v := vm.VM{}
	v.Init(program, params)

	b.ResetTimer()
	for n := 0; n < b.N; n++ {
		out, err = v.Run(program, params)
	}
	b.StopTimer()

	if err != nil {
		b.Fatal(err)
	}
	if !out.(bool) {
		b.Fail()
	}
}

func Benchmark_filter(b *testing.B) {
	params := make(map[string]interface{})
	params["max"] = 50

	program, err := jsexpr.Compile(`filter(1..100, {# > max})`, jsexpr.TypeCheck(params))
	if err != nil {
		b.Fatal(err)
	}

	virtualMachine := vm.VM{}
	virtualMachine.Init(program, params)
	for n := 0; n < b.N; n++ {
		_, err = virtualMachine.Run(program, params)
	}

	if err != nil {
		b.Fatal(err)
	}
}

func Benchmark_access(b *testing.B) {
	type Price struct {
		Value int `jsexpr:"value"`
	}
	type Env struct {
		Price Price `jsexpr:"price"`
	}

	program, err := jsexpr.Compile(`price.value > 0`, jsexpr.TypeCheck(Env{}))
	if err != nil {
		b.Fatal(err)
	}

	env := Env{Price: Price{Value: 1}}
	virtualMachine := vm.VM{}
	virtualMachine.Init(program, env)
	var out interface{}
	for n := 0; n < b.N; n++ {
		out, err = virtualMachine.Run(program, env)
	}

	if err != nil {
		b.Fatal(err)
	}
	assert.Nil(b, err)
	assert.Equal(b, true, out)
}

func Benchmark_accessMap(b *testing.B) {
	type Price struct {
		Value int `jsexpr:"value"`
	}
	env := map[string]interface{}{
		"price": Price{Value: 1},
	}

	program, err := jsexpr.Compile(`price.value > 0`, jsexpr.TypeCheck(env))
	if err != nil {
		b.Fatal(err)
	}

	for n := 0; n < b.N; n++ {
		_, err = vm.Run(program, env)
	}

	if err != nil {
		b.Fatal(err)
	}
}

func Benchmark_call(b *testing.B) {
	type Env struct {
		Fn func(string, string, string) bool `jsexpr:"fn"`
	}

	program, err := jsexpr.Compile(`fn("a", "b", "ab")`, jsexpr.TypeCheck(Env{}))
	// program, err := jsexpr.Compile(`fn("a", "b", "ab")`)
	if err != nil {
		b.Fatal(err)
	}

	env := Env{
		Fn: func(s1, s2, s3 string) bool {
			return s1+s2 == s3
		},
	}

	virtualMachine := vm.VM{}
	virtualMachine.Init(program, env)
	for n := 0; n < b.N; n++ {
		_, err = virtualMachine.Run(program, env)
	}

	if err != nil {
		b.Fatal(err)
	}
}

func Benchmark_callFast(b *testing.B) {
	type Env struct {
		Fn func(...interface{}) interface{} `jsexpr:"fn"`
	}

	program, err := jsexpr.Compile(`fn("a", "b", "ab")`, jsexpr.TypeCheck(Env{}))
	if err != nil {
		b.Fatal(err)
	}

	env := Env{
		Fn: func(s ...interface{}) interface{} {
			return s[0].(string)+s[1].(string) == s[2].(string)
		},
	}
	virtualMachine := vm.VM{}
	virtualMachine.Init(program, env)

	for n := 0; n < b.N; n++ {
		_, err = virtualMachine.Run(program, env)
	}

	if err != nil {
		b.Fatal(err)
	}
}

func Benchmark_callConstExpr(b *testing.B) {
	env := map[string]interface{}{
		"Fn": func(s ...interface{}) interface{} { return s[0].(string)+s[1].(string) == s[2].(string) },
	}

	program, err := jsexpr.Compile(`Fn("a", "b", "ab")`, jsexpr.TypeCheck(env), jsexpr.ConstExpr("Fn"))
	if err != nil {
		b.Fatal(err)
	}

	virtualMachine := vm.VM{}
	virtualMachine.Init(program, env)
	var out interface{}
	for n := 0; n < b.N; n++ {
		out, err = virtualMachine.Run(program, env)
	}

	if err != nil {
		b.Fatal(err)
	}
	if !out.(bool) {
		b.Fail()
	}
}

func Benchmark_largeStructAccess(b *testing.B) {
	type Env struct {
		Data  [1024 * 1024 * 10]byte `jsexpr:"data"`
		Field int                    `jsexpr:"field"`
	}

	program, err := jsexpr.Compile(`field > 0 && field > 1 && field < 20`, jsexpr.TypeCheck(Env{}))
	if err != nil {
		b.Fatal(err)
	}

	env := Env{Field: 21}
	virtualMachine := vm.VM{}
	virtualMachine.Init(program, env)

	for n := 0; n < b.N; n++ {
		_, err = virtualMachine.Run(program, &env)
	}

	if err != nil {
		b.Fatal(err)
	}
}

func Benchmark_largeNestedStructAccess(b *testing.B) {
	type Env struct {
		Inner struct {
			Data  [1024 * 1024 * 10]byte `jsexpr:"data"`
			Field int                    `jsexpr:"field"`
		} `jsexpr:"inner"`
	}

	program, err := jsexpr.Compile(`inner.field > 0 && inner.field > 1 && inner.field < 20`, jsexpr.TypeCheck(Env{}))
	if err != nil {
		b.Fatal(err)
	}

	env := Env{}
	env.Inner.Field = 21
	virtualMachine := vm.VM{}
	virtualMachine.Init(program, env)

	for n := 0; n < b.N; n++ {
		_, err = virtualMachine.Run(program, &env)
	}

	if err != nil {
		b.Fatal(err)
	}
}

func Benchmark_largeNestedArrayAccess(b *testing.B) {
	type Env struct {
		Data [1][1024 * 1024 * 10]byte `jsexpr:"data"`
	}

	program, err := jsexpr.Compile(`data[0][0] > 0`, jsexpr.TypeCheck(Env{}))
	if err != nil {
		b.Fatal(err)
	}

	env := Env{}
	virtualMachine := vm.VM{}
	virtualMachine.Init(program, env)
	for n := 0; n < b.N; n++ {
		_, err = virtualMachine.Run(program, &env)
	}

	if err != nil {
		b.Fatal(err)
	}
}

func createEnv() interface{} {
	type DirectFlightsDays struct {
		Start string `jsexpr:"start"`
		Days  string `jsexpr:"days"`
	}
	type RouteSegment struct {
		Origin                string             `jsexpr:"origin"`
		OriginName            string             `jsexpr:"originName"`
		Destination           string             `jsexpr:"destination"`
		DestinationName       string             `jsexpr:"destinationName"`
		Date                  string             `jsexpr:"date"`
		OriginCountry         string             `jsexpr:"originCountry"`
		DestinationCountry    string             `jsexpr:"destinationCountry"`
		TranslatedOrigin      string             `jsexpr:"translatedOrigin"`
		TranslatedDestination string             `jsexpr:"translatedDestination"`
		UserOrigin            string             `jsexpr:"userOrigin"`
		UserDestination       string             `jsexpr:"userDestination"`
		DirectFlightsDays     *DirectFlightsDays `jsexpr:"directFlightsDays"`
	}
	type Passengers struct {
		Adults   uint32 `jsexpr:"adults"`
		Children uint32 `jsexpr:"children"`
		Infants  uint32 `jsexpr:"infants"`
	}
	type UserAgentFeatures struct {
		Assisted     bool `jsexpr:"assisted"`
		TopPlacement bool `jsexpr:"topPlacement"`
		TourTickets  bool `jsexpr:"tourTickets"`
	}
	type SearchParamsEnv struct {
		Segments           []*RouteSegment    `jsexpr:"segments"`
		OriginCountry      string             `jsexpr:"originCountry"`
		DestinationCountry string             `jsexpr:"destinationCountry"`
		SearchDepth        int                `jsexpr:"searchDepth"`
		Passengers         *Passengers        `jsexpr:"passengers"`
		TripClass          string             `jsexpr:"tripClass"`
		UserIP             string             `jsexpr:"userIP"`
		KnowEnglish        bool               `jsexpr:"knowEnglish"`
		Market             string             `jsexpr:"market"`
		Marker             string             `jsexpr:"marker"`
		CleanMarker        string             `jsexpr:"cleanMarker"`
		Locale             string             `jsexpr:"locale"`
		ReferrerHost       string             `jsexpr:"referrerHost"`
		CountryCode        string             `jsexpr:"countryCode"`
		CurrencyCode       string             `jsexpr:"currencyCode"`
		IsOpenJaw          bool               `jsexpr:"isOpenJaw"`
		Os                 string             `jsexpr:"os"`
		OsVersion          string             `jsexpr:"osVersion"`
		AppVersion         string             `jsexpr:"appVersion"`
		IsAffiliate        bool               `jsexpr:"isAffiliate"`
		InitializedAt      int64              `jsexpr:"initializedAt"`
		Random             float32            `jsexpr:"random"`
		TravelPayoutsAPI   bool               `jsexpr:"travelPayoutsApi"`
		Features           *UserAgentFeatures `jsexpr:"features"`
		GateID             int32              `jsexpr:"gateID"`
		UserAgentDevice    string             `jsexpr:"userAgentDevice"`
		UserAgentType      string             `jsexpr:"userAgentType"`
		IsDesktop          bool               `jsexpr:"isDesktop"`
		IsMobile           bool               `jsexpr:"isMobile"`
	}
	type Env struct {
		SearchParamsEnv
	}
	return Env{
		SearchParamsEnv: SearchParamsEnv{
			Segments: []*RouteSegment{
				{
					Origin:      "VOG",
					Destination: "SHJ",
				},
				{
					Origin:      "SHJ",
					Destination: "VOG",
				},
			},
			OriginCountry:      "RU",
			DestinationCountry: "RU",
			SearchDepth:        44,
			Passengers:         &Passengers{1, 0, 0},
			TripClass:          "Y",
			UserIP:             "::1",
			KnowEnglish:        true,
			Market:             "ru",
			Marker:             "123456.direct",
			CleanMarker:        "123456",
			Locale:             "ru",
			ReferrerHost:       "www.aviasales.ru",
			CountryCode:        "",
			CurrencyCode:       "usd",
			IsOpenJaw:          false,
			Os:                 "",
			OsVersion:          "",
			AppVersion:         "",
			IsAffiliate:        true,
			InitializedAt:      1570788719,
			Random:             0.13497187,
			TravelPayoutsAPI:   false,
			Features:           &UserAgentFeatures{},
			GateID:             421,
			UserAgentDevice:    "DESKTOP",
			UserAgentType:      "WEB",
			IsDesktop:          true,
			IsMobile:           false,
		},
	}
}

func Benchmark_realWorld(b *testing.B) {
	env := createEnv()
	expression := `(userAgentDevice == 'DESKTOP') and ((originCountry == 'RU' or destinationCountry == 'RU') and market in ['ru', 'kz','by','uz','ua','az','am'])`
	program, err := jsexpr.Compile(expression, jsexpr.TypeCheck(env))
	if err != nil {
		b.Fatal(err)
	}

	virtualMachine := vm.VM{}
	virtualMachine.Init(program, env)
	var out interface{}
	for n := 0; n < b.N; n++ {
		out, err = virtualMachine.Run(program, env)
	}
	if err != nil {
		b.Fatal(err)
	}
	if !out.(bool) {
		b.Fail()
	}
}

func Benchmark_realWorld_reuseVm(b *testing.B) {
	env := createEnv()
	expression := `(userAgentDevice == 'DESKTOP') and ((originCountry == 'RU' or destinationCountry == 'RU') and market in ['ru', 'kz','by','uz','ua','az','am'])`
	program, err := jsexpr.Compile(expression, jsexpr.TypeCheck(env))
	if err != nil {
		b.Fatal(err)
	}

	var out interface{}
	v := vm.VM{}
	v.Init(program, env)

	for n := 0; n < b.N; n++ {
		out, err = v.Run(program, env)
	}

	if err != nil {
		b.Fatal(err)
	}
	if !out.(bool) {
		b.Fail()
	}
}

func Benchmark_realWorldInsane(b *testing.B) {
	env := createEnv()
	expression := `(userAgentDevice == 'DESKTOP') and (segments[0].origin in ['HKT','GOJ'] and segments[0].destination in ['HKT','GOJ'] or segments[0].origin in ['SKG','GOJ'] and segments[0].destination in ['SKG','GOJ'] or segments[0].origin in ['SSH','SVX'] and segments[0].destination in ['SSH','SVX'] or segments[0].origin in ['AYT','LED'] and segments[0].destination in ['AYT','LED'] or segments[0].origin in ['PUJ','KRR'] and segments[0].destination in ['PUJ','KRR'] or segments[0].origin in ['USM','CEK'] and segments[0].destination in ['USM','CEK'] or segments[0].origin in ['SHJ','LED'] and segments[0].destination in ['SHJ','LED'] or segments[0].origin in ['MOW','PRG'] and segments[0].destination in ['MOW','PRG'] or segments[0].origin in ['BKK','NOZ'] and segments[0].destination in ['BKK','NOZ'] or segments[0].origin in ['NHA','GOJ'] and segments[0].destination in ['NHA','GOJ'] or segments[0].origin in ['HRG','VOG'] and segments[0].destination in ['HRG','VOG'] or segments[0].origin in ['CFU','MSQ'] and segments[0].destination in ['CFU','MSQ'] or segments[0].origin in ['UFA','PUJ'] and segments[0].destination in ['UFA','PUJ'] or segments[0].origin in ['OMS','PUJ'] and segments[0].destination in ['OMS','PUJ'] or segments[0].origin in ['SKG','MSQ'] and segments[0].destination in ['SKG','MSQ'] or segments[0].origin in ['SSH','VOZ'] and segments[0].destination in ['SSH','VOZ'] or segments[0].origin in ['SSH','EGO'] and segments[0].destination in ['SSH','EGO'] or segments[0].origin in ['UUS','NHA'] and segments[0].destination in ['UUS','NHA'] or segments[0].origin in ['PUJ','MCX'] and segments[0].destination in ['PUJ','MCX'] or segments[0].origin in ['NHA','VVO'] and segments[0].destination in ['NHA','VVO'] or segments[0].origin in ['SKD','MOW'] and segments[0].destination in ['SKD','MOW'] or segments[0].origin in ['REN','NHA'] and segments[0].destination in ['REN','NHA'] or segments[0].origin in ['ASF','VRA'] and segments[0].destination in ['ASF','VRA'] or segments[0].origin in ['YKS','VRA'] and segments[0].destination in ['YKS','VRA'] or segments[0].origin in ['MOW','RIX'] and segments[0].destination in ['MOW','RIX'] or segments[0].origin in ['HER','IEV'] and segments[0].destination in ['HER','IEV'] or segments[0].origin in ['HRG','EGO'] and segments[0].destination in ['HRG','EGO'] or segments[0].origin in ['MOW','ATH'] and segments[0].destination in ['MOW','ATH'] or segments[0].origin in ['EGO','SSH'] and segments[0].destination in ['EGO','SSH'] or segments[0].origin in ['CEK','CUN'] and segments[0].destination in ['CEK','CUN'] or segments[0].origin in ['VAR','MOW'] and segments[0].destination in ['VAR','MOW'] or segments[0].origin in ['ASF','NHA'] and segments[0].destination in ['ASF','NHA'] or segments[0].origin in ['SKG','OVB'] and segments[0].destination in ['SKG','OVB'] or segments[0].origin in ['CUN','VOZ'] and segments[0].destination in ['CUN','VOZ'] or segments[0].origin in ['HRG','OVB'] and segments[0].destination in ['HRG','OVB'] or segments[0].origin in ['LED','VAR'] and segments[0].destination in ['LED','VAR'] or segments[0].origin in ['OMS','CUN'] and segments[0].destination in ['OMS','CUN'] or segments[0].origin in ['PUJ','NOZ'] and segments[0].destination in ['PUJ','NOZ'] or segments[0].origin in ['CUN','OMS'] and segments[0].destination in ['CUN','OMS'] or segments[0].origin in ['BAX','NHA'] and segments[0].destination in ['BAX','NHA'] or segments[0].origin in ['TDX','TJM'] and segments[0].destination in ['TDX','TJM'] or segments[0].origin in ['BKK','YKS'] and segments[0].destination in ['BKK','YKS'] or segments[0].origin in ['PUJ','MRV'] and segments[0].destination in ['PUJ','MRV'] or segments[0].origin in ['KUF','MOW'] and segments[0].destination in ['KUF','MOW'] or segments[0].origin in ['NHA','YKS'] and segments[0].destination in ['NHA','YKS'] or segments[0].origin in ['UFA','CUN'] and segments[0].destination in ['UFA','CUN'] or segments[0].origin in ['MIR','MOW'] and segments[0].destination in ['MIR','MOW'] or segments[0].origin in ['OVB','PUJ'] and segments[0].destination in ['OVB','PUJ'] or segments[0].origin in ['SGN','KJA'] and segments[0].destination in ['SGN','KJA'] or segments[0].origin in ['UTP','CEK'] and segments[0].destination in ['UTP','CEK'] or segments[0].origin in ['SKG','IEV'] and segments[0].destination in ['SKG','IEV'] or segments[0].origin in ['PKC','MOW'] and segments[0].destination in ['PKC','MOW'] or segments[0].origin in ['NHA','OGZ'] and segments[0].destination in ['NHA','OGZ'] or segments[0].origin in ['USM','UFA'] and segments[0].destination in ['USM','UFA'] or segments[0].origin in ['KGD','VRA'] and segments[0].destination in ['KGD','VRA'] or segments[0].origin in ['TDX','KZN'] and segments[0].destination in ['TDX','KZN'] or segments[0].origin in ['KRR','CUN'] and segments[0].destination in ['KRR','CUN'] or segments[0].origin in ['DXB','PEE'] and segments[0].destination in ['DXB','PEE'] or segments[0].origin in ['AER','KUF'] and segments[0].destination in ['AER','KUF'] or segments[0].origin in ['REN','SSH'] and segments[0].destination in ['REN','SSH'] or segments[0].origin in ['HKT','NJC'] and segments[0].destination in ['HKT','NJC'] or segments[0].origin in ['AER','CUN'] and segments[0].destination in ['AER','CUN'] or segments[0].origin in ['ETH','SVX'] and segments[0].destination in ['ETH','SVX'] or segments[0].origin in ['SSH','CEK'] and segments[0].destination in ['SSH','CEK'] or segments[0].origin in ['BKK','UFA'] and segments[0].destination in ['BKK','UFA'] or segments[0].origin in ['SVX','SKG'] and segments[0].destination in ['SVX','SKG'] or segments[0].origin in ['BKK','VOG'] and segments[0].destination in ['BKK','VOG'] or segments[0].origin in ['SKG','MOW'] and segments[0].destination in ['SKG','MOW'] or segments[0].origin in ['NHA','NOZ'] and segments[0].destination in ['NHA','NOZ'] or segments[0].origin in ['YKS','OVB'] and segments[0].destination in ['YKS','OVB'] or segments[0].origin in ['UFA','VRA'] and segments[0].destination in ['UFA','VRA'] or segments[0].origin in ['MOW','TCI'] and segments[0].destination in ['MOW','TCI'] or segments[0].origin in ['ASF','PUJ'] and segments[0].destination in ['ASF','PUJ'] or segments[0].origin in ['GOJ','CUN'] and segments[0].destination in ['GOJ','CUN'] or segments[0].origin in ['ASF','CUN'] and segments[0].destination in ['ASF','CUN'] or segments[0].origin in ['SGN','CEK'] and segments[0].destination in ['SGN','CEK'] or segments[0].origin in ['TJM','SSH'] and segments[0].destination in ['TJM','SSH'] or segments[0].origin in ['UTP','KZN'] and segments[0].destination in ['UTP','KZN'] or segments[0].origin in ['HRG','REN'] and segments[0].destination in ['HRG','REN'] or segments[0].origin in ['HKT','KJA'] and segments[0].destination in ['HKT','KJA'] or segments[0].origin in ['BEG','MOW'] and segments[0].destination in ['BEG','MOW'] or segments[0].origin in ['OMS','SSH'] and segments[0].destination in ['OMS','SSH'] or segments[0].origin in ['MSQ','SKG'] and segments[0].destination in ['MSQ','SKG'] or segments[0].origin in ['BKK','HTA'] and segments[0].destination in ['BKK','HTA'] or segments[0].origin in ['TDX','PEE'] and segments[0].destination in ['TDX','PEE'] or segments[0].origin in ['SKG','MRV'] and segments[0].destination in ['SKG','MRV'] or segments[0].origin in ['SGN','OVB'] and segments[0].destination in ['SGN','OVB'] or segments[0].origin in ['SVX','HRG'] and segments[0].destination in ['SVX','HRG'] or segments[0].origin in ['HKT','AER'] and segments[0].destination in ['HKT','AER'] or segments[0].origin in ['CEE','CUN'] and segments[0].destination in ['CEE','CUN'] or segments[0].origin in ['NHA','SVX'] and segments[0].destination in ['NHA','SVX'] or segments[0].origin in ['CUN','GOJ'] and segments[0].destination in ['CUN','GOJ'] or segments[0].origin in ['MOW','OGZ'] and segments[0].destination in ['MOW','OGZ'] or segments[0].origin in ['SCW','SSH'] and segments[0].destination in ['SCW','SSH'] or segments[0].origin in ['PUJ','PEE'] and segments[0].destination in ['PUJ','PEE'] or segments[0].origin in ['CUN','ASF'] and segments[0].destination in ['CUN','ASF'] or segments[0].origin in ['AQJ','SVX'] and segments[0].destination in ['AQJ','SVX'] or segments[0].origin in ['VRA','IKT'] and segments[0].destination in ['VRA','IKT'] or segments[0].origin in ['SHJ','SVX'] and segments[0].destination in ['SHJ','SVX'] or segments[0].origin in ['NBC','VRA'] and segments[0].destination in ['NBC','VRA'] or segments[0].origin in ['HTA','CUN'] and segments[0].destination in ['HTA','CUN'] or segments[0].origin in ['MOW','TOF'] and segments[0].destination in ['MOW','TOF'] or segments[0].origin in ['NJC','CUN'] and segments[0].destination in ['NJC','CUN'] or segments[0].origin in ['CUN','NOZ'] and segments[0].destination in ['CUN','NOZ'] or segments[0].origin in ['BTK','NHA'] and segments[0].destination in ['BTK','NHA'] or segments[0].origin in ['PUJ','OMS'] and segments[0].destination in ['PUJ','OMS'] or segments[0].origin in ['HTA','OVB'] and segments[0].destination in ['HTA','OVB'] or segments[0].origin in ['AQJ','KZN'] and segments[0].destination in ['AQJ','KZN'] or segments[0].origin in ['DXB','VOZ'] and segments[0].destination in ['DXB','VOZ'] or segments[0].origin in ['NHA','PEE'] and segments[0].destination in ['NHA','PEE'] or segments[0].origin in ['HKT','OGZ'] and segments[0].destination in ['HKT','OGZ'] or segments[0].origin in ['KLV','MOW'] and segments[0].destination in ['KLV','MOW'] or segments[0].origin in ['MRV','SKG'] and segments[0].destination in ['MRV','SKG'] or segments[0].origin in ['SKG','LED'] and segments[0].destination in ['SKG','LED'] or segments[0].origin in ['AQJ','MOW'] and segments[0].destination in ['AQJ','MOW'] or segments[0].origin in ['MOW','NHA'] and segments[0].destination in ['MOW','NHA'] or segments[0].origin in ['ARH','HRG'] and segments[0].destination in ['ARH','HRG'] or segments[0].origin in ['SGN','AER'] and segments[0].destination in ['SGN','AER'] or segments[0].origin in ['VRA','MCX'] and segments[0].destination in ['VRA','MCX'] or segments[0].origin in ['BKK','OVB'] and segments[0].destination in ['BKK','OVB'] or segments[0].origin in ['AYT','UFA'] and segments[0].destination in ['AYT','UFA'] or segments[0].origin in ['SGN','NOZ'] and segments[0].destination in ['SGN','NOZ'] or segments[0].origin in ['SGN','NBC'] and segments[0].destination in ['SGN','NBC'] or segments[0].origin in ['MOW','BEG'] and segments[0].destination in ['MOW','BEG'] or segments[0].origin in ['TDX','BQS'] and segments[0].destination in ['TDX','BQS'] or segments[0].origin in ['KRR','NHA'] and segments[0].destination in ['KRR','NHA'] or segments[0].origin in ['NHA','SGC'] and segments[0].destination in ['NHA','SGC'] or segments[0].origin in ['NHA','UFA'] and segments[0].destination in ['NHA','UFA'] or segments[0].origin in ['NHA','ARH'] and segments[0].destination in ['NHA','ARH'] or segments[0].origin in ['EGO','VRA'] and segments[0].destination in ['EGO','VRA'] or segments[0].origin in ['BCN','MOW'] and segments[0].destination in ['BCN','MOW'] or segments[0].origin in ['TDX','ROV'] and segments[0].destination in ['TDX','ROV'] or segments[0].origin in ['TSN','MOW'] and segments[0].destination in ['TSN','MOW'] or segments[0].origin in ['GOJ','HRG'] and segments[0].destination in ['GOJ','HRG'] or segments[0].origin in ['BKK','KZN'] and segments[0].destination in ['BKK','KZN'] or segments[0].origin in ['NHA','ROV'] and segments[0].destination in ['NHA','ROV'] or segments[0].origin in ['DXB','KJA'] and segments[0].destination in ['DXB','KJA'] or segments[0].origin in ['PEE','AER'] and segments[0].destination in ['PEE','AER'] or segments[0].origin in ['DXB','CEK'] and segments[0].destination in ['DXB','CEK'] or segments[0].origin in ['PUJ','ASF'] and segments[0].destination in ['PUJ','ASF'] or segments[0].origin in ['KBV','OVB'] and segments[0].destination in ['KBV','OVB'] or segments[0].origin in ['MOW','EVN'] and segments[0].destination in ['MOW','EVN'] or segments[0].origin in ['IKT','CUN'] and segments[0].destination in ['IKT','CUN'] or segments[0].origin in ['KGD','HRG'] and segments[0].destination in ['KGD','HRG'] or segments[0].origin in ['KBV','PEE'] and segments[0].destination in ['KBV','PEE'] or segments[0].origin in ['VOG','VRA'] and segments[0].destination in ['VOG','VRA'] or segments[0].origin in ['MOW','HKT'] and segments[0].destination in ['MOW','HKT'] or segments[0].origin in ['NHA','ASF'] and segments[0].destination in ['NHA','ASF'] or segments[0].origin in ['LED','SVX'] and segments[0].destination in ['LED','SVX'] or segments[0].origin in ['AAQ','CUN'] and segments[0].destination in ['AAQ','CUN'] or segments[0].origin in ['BKK','KEJ'] and segments[0].destination in ['BKK','KEJ'] or segments[0].origin in ['BKK','BQS'] and segments[0].destination in ['BKK','BQS'] or segments[0].origin in ['DXB','IKT'] and segments[0].destination in ['DXB','IKT'] or segments[0].origin in ['SSH','TJM'] and segments[0].destination in ['SSH','TJM'] or segments[0].origin in ['PUJ','ROV'] and segments[0].destination in ['PUJ','ROV'] or segments[0].origin in ['AER','SVX'] and segments[0].destination in ['AER','SVX'] or segments[0].origin in ['UFA','ETH'] and segments[0].destination in ['UFA','ETH'] or segments[0].origin in ['BKK','KUF'] and segments[0].destination in ['BKK','KUF'] or segments[0].origin in ['BKK','VVO'] and segments[0].destination in ['BKK','VVO'] or segments[0].origin in ['HKT','OVB'] and segments[0].destination in ['HKT','OVB'] or segments[0].origin in ['ZTH','LED'] and segments[0].destination in ['ZTH','LED'] or segments[0].origin in ['KZN','NHA'] and segments[0].destination in ['KZN','NHA'] or segments[0].origin in ['VRA','BAX'] and segments[0].destination in ['VRA','BAX'] or segments[0].origin in ['RTW','NHA'] and segments[0].destination in ['RTW','NHA'] or segments[0].origin in ['SKG','DNK'] and segments[0].destination in ['SKG','DNK'] or segments[0].origin in ['SGN','VOG'] and segments[0].destination in ['SGN','VOG'] or segments[0].origin in ['KBV','VVO'] and segments[0].destination in ['KBV','VVO'] or segments[0].origin in ['IEV','CFU'] and segments[0].destination in ['IEV','CFU'] or segments[0].origin in ['PUJ','TOF'] and segments[0].destination in ['PUJ','TOF'] or segments[0].origin in ['HKT','KEJ'] and segments[0].destination in ['HKT','KEJ'] or segments[0].origin in ['PUJ','NJC'] and segments[0].destination in ['PUJ','NJC'] or segments[0].origin in ['PEE','CUN'] and segments[0].destination in ['PEE','CUN'] or segments[0].origin in ['HKT','TJM'] and segments[0].destination in ['HKT','TJM'] or segments[0].origin in ['ETH','KZN'] and segments[0].destination in ['ETH','KZN'] or segments[0].origin in ['MCX','CUN'] and segments[0].destination in ['MCX','CUN'] or segments[0].origin in ['HRG','KUF'] and segments[0].destination in ['HRG','KUF'] or segments[0].origin in ['VRA','VOG'] and segments[0].destination in ['VRA','VOG'] or segments[0].origin in ['SVX','CUN'] and segments[0].destination in ['SVX','CUN'] or segments[0].origin in ['VRA','EGO'] and segments[0].destination in ['VRA','EGO'] or segments[0].origin in ['ROV','CUN'] and segments[0].destination in ['ROV','CUN'] or segments[0].origin in ['KJA','VRA'] and segments[0].destination in ['KJA','VRA'] or segments[0].origin in ['VRA','PEE'] and segments[0].destination in ['VRA','PEE'] or segments[0].origin in ['MOW','SKD'] and segments[0].destination in ['MOW','SKD'] or segments[0].origin in ['POP','ROV'] and segments[0].destination in ['POP','ROV'] or segments[0].origin in ['AYT','KZN'] and segments[0].destination in ['AYT','KZN'] or segments[0].origin in ['ETH','REN'] and segments[0].destination in ['ETH','REN'] or segments[0].origin in ['ETH','LED'] and segments[0].destination in ['ETH','LED'] or segments[0].origin in ['CEK','ETH'] and segments[0].destination in ['CEK','ETH'] or segments[0].origin in ['NHA','VOZ'] and segments[0].destination in ['NHA','VOZ'] or segments[0].origin in ['SVX','AER'] and segments[0].destination in ['SVX','AER'] or segments[0].origin in ['FEG','MOW'] and segments[0].destination in ['FEG','MOW'] or segments[0].origin in ['VRA','KZN'] and segments[0].destination in ['VRA','KZN'] or segments[0].origin in ['USM','PEE'] and segments[0].destination in ['USM','PEE'] or segments[0].origin in ['VVO','MOW'] and segments[0].destination in ['VVO','MOW'] or segments[0].origin in ['SGN','KEJ'] and segments[0].destination in ['SGN','KEJ'] or segments[0].origin in ['DXB','AER'] and segments[0].destination in ['DXB','AER'] or segments[0].origin in ['MOW','VOG'] and segments[0].destination in ['MOW','VOG'] or segments[0].origin in ['SGN','YKS'] and segments[0].destination in ['SGN','YKS'] or segments[0].origin in ['VRA','NJC'] and segments[0].destination in ['VRA','NJC'] or segments[0].origin in ['VOG','PUJ'] and segments[0].destination in ['VOG','PUJ'] or segments[0].origin in ['HKT','MOW'] and segments[0].destination in ['HKT','MOW'] or segments[0].origin in ['VOG','SKG'] and segments[0].destination in ['VOG','SKG'] or segments[0].origin in ['OVB','YKS'] and segments[0].destination in ['OVB','YKS'] or segments[0].origin in ['SGC','SSH'] and segments[0].destination in ['SGC','SSH'] or segments[0].origin in ['VOZ','NHA'] and segments[0].destination in ['VOZ','NHA'] or segments[0].origin in ['CUN','NBC'] and segments[0].destination in ['CUN','NBC'] or segments[0].origin in ['KZN','SSH'] and segments[0].destination in ['KZN','SSH'] or segments[0].origin in ['HER','MOW'] and segments[0].destination in ['HER','MOW'] or segments[0].origin in ['TDX','UFA'] and segments[0].destination in ['TDX','UFA'] or segments[0].origin in ['KZN','ETH'] and segments[0].destination in ['KZN','ETH'] or segments[0].origin in ['ABA','CUN'] and segments[0].destination in ['ABA','CUN'] or segments[0].origin in ['PEE','NHA'] and segments[0].destination in ['PEE','NHA'] or segments[0].origin in ['CUN','TOF'] and segments[0].destination in ['CUN','TOF'] or segments[0].origin in ['TJM','HRG'] and segments[0].destination in ['TJM','HRG'] or segments[0].origin in ['EGO','HRG'] and segments[0].destination in ['EGO','HRG'] or segments[0].origin in ['GOJ','SSH'] and segments[0].destination in ['GOJ','SSH'] or segments[0].origin in ['HKT','HTA'] and segments[0].destination in ['HKT','HTA'] or segments[0].origin in ['MOW','ETH'] and segments[0].destination in ['MOW','ETH'] or segments[0].origin in ['OGZ','VRA'] and segments[0].destination in ['OGZ','VRA'] or segments[0].origin in ['HKT','NBC'] and segments[0].destination in ['HKT','NBC'] or segments[0].origin in ['GPA','MSQ'] and segments[0].destination in ['GPA','MSQ'] or segments[0].origin in ['SGN','TOF'] and segments[0].destination in ['SGN','TOF'] or segments[0].origin in ['HKT','MCX'] and segments[0].destination in ['HKT','MCX'] or segments[0].origin in ['KRR','VRA'] and segments[0].destination in ['KRR','VRA'] or segments[0].origin in ['ROV','PUJ'] and segments[0].destination in ['ROV','PUJ'] or segments[0].origin in ['CEE','VRA'] and segments[0].destination in ['CEE','VRA'] or segments[0].origin in ['TJM','NHA'] and segments[0].destination in ['TJM','NHA'] or segments[0].origin in ['RTW','CUN'] and segments[0].destination in ['RTW','CUN'] or segments[0].origin in ['AER','KZN'] and segments[0].destination in ['AER','KZN'] or segments[0].origin in ['MRV','ETH'] and segments[0].destination in ['MRV','ETH'] or segments[0].origin in ['SGN','VOZ'] and segments[0].destination in ['SGN','VOZ'] or segments[0].origin in ['USM','BQS'] and segments[0].destination in ['USM','BQS'] or segments[0].origin in ['USM','SGC'] and segments[0].destination in ['USM','SGC'] or segments[0].origin in ['HER','SVX'] and segments[0].destination in ['HER','SVX'] or segments[0].origin in ['DXB','KZN'] and segments[0].destination in ['DXB','KZN'] or segments[0].origin in ['TDX','KEJ'] and segments[0].destination in ['TDX','KEJ'] or segments[0].origin in ['HRG','SGC'] and segments[0].destination in ['HRG','SGC'] or segments[0].origin in ['SOF','LED'] and segments[0].destination in ['SOF','LED'] or segments[0].origin in ['DXB','UFA'] and segments[0].destination in ['DXB','UFA'] or segments[0].origin in ['EVN','MOW'] and segments[0].destination in ['EVN','MOW'] or segments[0].origin in ['HKT','LED'] and segments[0].destination in ['HKT','LED'] or segments[0].origin in ['SGN','NJC'] and segments[0].destination in ['SGN','NJC'] or segments[0].origin in ['SHJ','KUF'] and segments[0].destination in ['SHJ','KUF'] or segments[0].origin in ['AQJ','LED'] and segments[0].destination in ['AQJ','LED'] or segments[0].origin in ['HRG','GOJ'] and segments[0].destination in ['HRG','GOJ'] or segments[0].origin in ['PRG','LED'] and segments[0].destination in ['PRG','LED'] or segments[0].origin in ['NOZ','NHA'] and segments[0].destination in ['NOZ','NHA'] or segments[0].origin in ['ARH','SSH'] and segments[0].destination in ['ARH','SSH'] or segments[0].origin in ['SSH','REN'] and segments[0].destination in ['SSH','REN'] or segments[0].origin in ['AYT','GOJ'] and segments[0].destination in ['AYT','GOJ'] or segments[0].origin in ['ATH','MSQ'] and segments[0].destination in ['ATH','MSQ'] or segments[0].origin in ['MOW','VAR'] and segments[0].destination in ['MOW','VAR'] or segments[0].origin in ['HER','LED'] and segments[0].destination in ['HER','LED'] or segments[0].origin in ['SIP','KJA'] and segments[0].destination in ['SIP','KJA'] or segments[0].origin in ['TJM','CUN'] and segments[0].destination in ['TJM','CUN'] or segments[0].origin in ['PUJ','LED'] and segments[0].destination in ['PUJ','LED'] or segments[0].origin in ['BKK','SGC'] and segments[0].destination in ['BKK','SGC'] or segments[0].origin in ['PUJ','KEJ'] and segments[0].destination in ['PUJ','KEJ'] or segments[0].origin in ['BKK','KJA'] and segments[0].destination in ['BKK','KJA'] or segments[0].origin in ['DXB','VOG'] and segments[0].destination in ['DXB','VOG'] or segments[0].origin in ['PUJ','KJA'] and segments[0].destination in ['PUJ','KJA'] or segments[0].origin in ['RMI','MOW'] and segments[0].destination in ['RMI','MOW'] or segments[0].origin in ['USM','KEJ'] and segments[0].destination in ['USM','KEJ'] or segments[0].origin in ['MOW','RVN'] and segments[0].destination in ['MOW','RVN'] or segments[0].origin in ['VRA','AER'] and segments[0].destination in ['VRA','AER'] or segments[0].origin in ['SGN','VVO'] and segments[0].destination in ['SGN','VVO'] or segments[0].origin in ['SIP','MOW'] and segments[0].destination in ['SIP','MOW'] or segments[0].origin in ['ETH','MRV'] and segments[0].destination in ['ETH','MRV'] or segments[0].origin in ['VRA','MRV'] and segments[0].destination in ['VRA','MRV'] or segments[0].origin in ['ROV','MOW'] and segments[0].destination in ['ROV','MOW'] or segments[0].origin in ['KBV','TJM'] and segments[0].destination in ['KBV','TJM'] or segments[0].origin in ['PUJ','VOZ'] and segments[0].destination in ['PUJ','VOZ'] or segments[0].origin in ['LED','AER'] and segments[0].destination in ['LED','AER'] or segments[0].origin in ['AER','VRA'] and segments[0].destination in ['AER','VRA'] or segments[0].origin in ['CUN','SVX'] and segments[0].destination in ['CUN','SVX'] or segments[0].origin in ['HKT','ROV'] and segments[0].destination in ['HKT','ROV'] or segments[0].origin in ['KUF','NHA'] and segments[0].destination in ['KUF','NHA'] or segments[0].origin in ['KGD','SKG'] and segments[0].destination in ['KGD','SKG'] or segments[0].origin in ['DXB','YKS'] and segments[0].destination in ['DXB','YKS'] or segments[0].origin in ['AER','PEE'] and segments[0].destination in ['AER','PEE'] or segments[0].origin in ['ROV','CFU'] and segments[0].destination in ['ROV','CFU'] or segments[0].origin in ['VOG','CUN'] and segments[0].destination in ['VOG','CUN'] or segments[0].origin in ['PUJ','KZN'] and segments[0].destination in ['PUJ','KZN'] or segments[0].origin in ['MOW','SZG'] and segments[0].destination in ['MOW','SZG'] or segments[0].origin in ['GDX','MOW'] and segments[0].destination in ['GDX','MOW'] or segments[0].origin in ['HKT','VOG'] and segments[0].destination in ['HKT','VOG'] or segments[0].origin in ['BOJ','MOW'] and segments[0].destination in ['BOJ','MOW'] or segments[0].origin in ['OVB','HTA'] and segments[0].destination in ['OVB','HTA'] or segments[0].origin in ['BKK','EGO'] and segments[0].destination in ['BKK','EGO'] or segments[0].origin in ['ETH','KUF'] and segments[0].destination in ['ETH','KUF'] or segments[0].origin in ['HRG','ARH'] and segments[0].destination in ['HRG','ARH'] or segments[0].origin in ['MOW','KGD'] and segments[0].destination in ['MOW','KGD'] or segments[0].origin in ['HRG','CEK'] and segments[0].destination in ['HRG','CEK'] or segments[0].origin in ['LED','HER'] and segments[0].destination in ['LED','HER'] or segments[0].origin in ['USM','IKT'] and segments[0].destination in ['USM','IKT'] or segments[0].origin in ['CUN','TJM'] and segments[0].destination in ['CUN','TJM'] or segments[0].origin in ['NHA','UUS'] and segments[0].destination in ['NHA','UUS'] or segments[0].origin in ['NHA','KZN'] and segments[0].destination in ['NHA','KZN'] or segments[0].origin in ['NBC','HRG'] and segments[0].destination in ['NBC','HRG'] or segments[0].origin in ['SKG','SVX'] and segments[0].destination in ['SKG','SVX'] or segments[0].origin in ['HRG','UFA'] and segments[0].destination in ['HRG','UFA'] or segments[0].origin in ['TDX','MOW'] and segments[0].destination in ['TDX','MOW'] or segments[0].origin in ['LED','SKG'] and segments[0].destination in ['LED','SKG'] or segments[0].origin in ['SGN','SVX'] and segments[0].destination in ['SGN','SVX'] or segments[0].origin in ['CUN','AER'] and segments[0].destination in ['CUN','AER'] or segments[0].origin in ['MOW','KUT'] and segments[0].destination in ['MOW','KUT'] or segments[0].origin in ['VRN','KRR'] and segments[0].destination in ['VRN','KRR'] or segments[0].origin in ['MSQ','ATH'] and segments[0].destination in ['MSQ','ATH'] or segments[0].origin in ['PUJ','BAX'] and segments[0].destination in ['PUJ','BAX'] or segments[0].origin in ['KEJ','CUN'] and segments[0].destination in ['KEJ','CUN'] or segments[0].origin in ['KUF','PUJ'] and segments[0].destination in ['KUF','PUJ'] or segments[0].origin in ['VRA','KUF'] and segments[0].destination in ['VRA','KUF'] or segments[0].origin in ['LED','HRG'] and segments[0].destination in ['LED','HRG'] or segments[0].origin in ['BKK','ASF'] and segments[0].destination in ['BKK','ASF'] or segments[0].origin in ['IEV','HER'] and segments[0].destination in ['IEV','HER'] or segments[0].origin in ['SHJ','ROV'] and segments[0].destination in ['SHJ','ROV'] or segments[0].origin in ['KUT','MOW'] and segments[0].destination in ['KUT','MOW'] or segments[0].origin in ['HKT','KRR'] and segments[0].destination in ['HKT','KRR'] or segments[0].origin in ['AYT','MOW'] and segments[0].destination in ['AYT','MOW'] or segments[0].origin in ['VRA','MOW'] and segments[0].destination in ['VRA','MOW'] or segments[0].origin in ['SCW','PUJ'] and segments[0].destination in ['SCW','PUJ'] or segments[0].origin in ['MOW','TAS'] and segments[0].destination in ['MOW','TAS'] or segments[0].origin in ['IEV','SKG'] and segments[0].destination in ['IEV','SKG'] or segments[0].origin in ['LED','BOJ'] and segments[0].destination in ['LED','BOJ'] or segments[0].origin in ['HKT','SVX'] and segments[0].destination in ['HKT','SVX'] or segments[0].origin in ['BKK','SVX'] and segments[0].destination in ['BKK','SVX'] or segments[0].origin in ['SGN','MOW'] and segments[0].destination in ['SGN','MOW'] or segments[0].origin in ['SVX','ETH'] and segments[0].destination in ['SVX','ETH'] or segments[0].origin in ['SSH','PEE'] and segments[0].destination in ['SSH','PEE'] or segments[0].origin in ['NHA','KUF'] and segments[0].destination in ['NHA','KUF'] or segments[0].origin in ['SSH','KUF'] and segments[0].destination in ['SSH','KUF'] or segments[0].origin in ['DXB','MOW'] and segments[0].destination in ['DXB','MOW'] or segments[0].origin in ['PUJ','YKS'] and segments[0].destination in ['PUJ','YKS'] or segments[0].origin in ['SSH','ARH'] and segments[0].destination in ['SSH','ARH'] or segments[0].origin in ['AUH','MOW'] and segments[0].destination in ['AUH','MOW'] or segments[0].origin in ['UTP','IKT'] and segments[0].destination in ['UTP','IKT'] or segments[0].origin in ['KRR','SSH'] and segments[0].destination in ['KRR','SSH'] or segments[0].origin in ['HRG','KZN'] and segments[0].destination in ['HRG','KZN'] or segments[0].origin in ['BKK','ROV'] and segments[0].destination in ['BKK','ROV'] or segments[0].origin in ['CEK','PUJ'] and segments[0].destination in ['CEK','PUJ'] or segments[0].origin in ['SGN','KGD'] and segments[0].destination in ['SGN','KGD'] or segments[0].origin in ['KEJ','PUJ'] and segments[0].destination in ['KEJ','PUJ'] or segments[0].origin in ['HKT','SCW'] and segments[0].destination in ['HKT','SCW'] or segments[0].origin in ['BKK','KGD'] and segments[0].destination in ['BKK','KGD'] or segments[0].origin in ['HKT','SGC'] and segments[0].destination in ['HKT','SGC'] or segments[0].origin in ['REN','HRG'] and segments[0].destination in ['REN','HRG'] or segments[0].origin in ['SKG','TSE'] and segments[0].destination in ['SKG','TSE'] or segments[0].origin in ['BKK','PKC'] and segments[0].destination in ['BKK','PKC'] or segments[0].origin in ['VRA','KJA'] and segments[0].destination in ['VRA','KJA'] or segments[0].origin in ['SCW','CUN'] and segments[0].destination in ['SCW','CUN'] or segments[0].origin in ['SKG','KZN'] and segments[0].destination in ['SKG','KZN'] or segments[0].origin in ['MOW','GRV'] and segments[0].destination in ['MOW','GRV'] or segments[0].origin in ['HRG','NBC'] and segments[0].destination in ['HRG','NBC'] or segments[0].origin in ['SCW','VRA'] and segments[0].destination in ['SCW','VRA'] or segments[0].origin in ['UFA','HRG'] and segments[0].destination in ['UFA','HRG'] or segments[0].origin in ['EGO','CUN'] and segments[0].destination in ['EGO','CUN'] or segments[0].origin in ['KUF','HRG'] and segments[0].destination in ['KUF','HRG'] or segments[0].origin in ['CUN','ROV'] and segments[0].destination in ['CUN','ROV'] or segments[0].origin in ['KBV','KEJ'] and segments[0].destination in ['KBV','KEJ'] or segments[0].origin in ['NHA','IKT'] and segments[0].destination in ['NHA','IKT'] or segments[0].origin in ['SSH','KRR'] and segments[0].destination in ['SSH','KRR'] or segments[0].origin in ['CFU','MOW'] and segments[0].destination in ['CFU','MOW'] or segments[0].origin in ['MSQ','GPA'] and segments[0].destination in ['MSQ','GPA'] or segments[0].origin in ['ZTH','MOW'] and segments[0].destination in ['ZTH','MOW'] or segments[0].origin in ['AER','KJA'] and segments[0].destination in ['AER','KJA'] or segments[0].origin in ['MOW','CFU'] and segments[0].destination in ['MOW','CFU'] or segments[0].origin in ['BKK','SCW'] and segments[0].destination in ['BKK','SCW'] or segments[0].origin in ['PUJ','OGZ'] and segments[0].destination in ['PUJ','OGZ'] or segments[0].origin in ['AMM','MOW'] and segments[0].destination in ['AMM','MOW'] or segments[0].origin in ['OVB','TOF'] and segments[0].destination in ['OVB','TOF'] or segments[0].origin in ['SGN','KZN'] and segments[0].destination in ['SGN','KZN'] or segments[0].origin in ['VOG','AER'] and segments[0].destination in ['VOG','AER'] or segments[0].origin in ['VRA','SVX'] and segments[0].destination in ['VRA','SVX'] or segments[0].origin in ['DXB','SVX'] and segments[0].destination in ['DXB','SVX'] or segments[0].origin in ['HKT','BQS'] and segments[0].destination in ['HKT','BQS'] or segments[0].origin in ['PUJ','EGO'] and segments[0].destination in ['PUJ','EGO'] or segments[0].origin in ['DXB','LED'] and segments[0].destination in ['DXB','LED'] or segments[0].origin in ['ETH','MOW'] and segments[0].destination in ['ETH','MOW'] or segments[0].origin in ['MOW','KJA'] and segments[0].destination in ['MOW','KJA'] or segments[0].origin in ['IKT','MOW'] and segments[0].destination in ['IKT','MOW'] or segments[0].origin in ['KBV','ROV'] and segments[0].destination in ['KBV','ROV'] or segments[0].origin in ['BKK','REN'] and segments[0].destination in ['BKK','REN'] or segments[0].origin in ['HKT','PEE'] and segments[0].destination in ['HKT','PEE'] or segments[0].origin in ['SVX','VRA'] and segments[0].destination in ['SVX','VRA'] or segments[0].origin in ['BKK','AER'] and segments[0].destination in ['BKK','AER'] or segments[0].origin in ['ETH','ROV'] and segments[0].destination in ['ETH','ROV'] or segments[0].origin in ['SGN','SCW'] and segments[0].destination in ['SGN','SCW'] or segments[0].origin in ['SIP','KUF'] and segments[0].destination in ['SIP','KUF'] or segments[0].origin in ['CEK','NHA'] and segments[0].destination in ['CEK','NHA'] or segments[0].origin in ['AQJ','KRR'] and segments[0].destination in ['AQJ','KRR'] or segments[0].origin in ['KBV','MOW'] and segments[0].destination in ['KBV','MOW'] or segments[0].origin in ['BHK','MOW'] and segments[0].destination in ['BHK','MOW'] or segments[0].origin in ['BKK','PEE'] and segments[0].destination in ['BKK','PEE'] or segments[0].origin in ['MOW','BAX'] and segments[0].destination in ['MOW','BAX'] or segments[0].origin in ['GPA','MOW'] and segments[0].destination in ['GPA','MOW'] or segments[0].origin in ['RIX','MOW'] and segments[0].destination in ['RIX','MOW'] or segments[0].origin in ['DXB','NBC'] and segments[0].destination in ['DXB','NBC'] or segments[0].origin in ['PUJ','OVB'] and segments[0].destination in ['PUJ','OVB'] or segments[0].origin in ['ETH','CEK'] and segments[0].destination in ['ETH','CEK'] or segments[0].origin in ['KRR','ETH'] and segments[0].destination in ['KRR','ETH'] or segments[0].origin in ['HKT','UUD'] and segments[0].destination in ['HKT','UUD'] or segments[0].origin in ['TOF','VRA'] and segments[0].destination in ['TOF','VRA'] or segments[0].origin in ['MOW','SKG'] and segments[0].destination in ['MOW','SKG'] or segments[0].origin in ['BTK','OVB'] and segments[0].destination in ['BTK','OVB'] or segments[0].origin in ['KRR','LCA'] and segments[0].destination in ['KRR','LCA'] or segments[0].origin in ['OGZ','CUN'] and segments[0].destination in ['OGZ','CUN'] or segments[0].origin in ['PUJ','KGD'] and segments[0].destination in ['PUJ','KGD'] or segments[0].origin in ['USM','OVB'] and segments[0].destination in ['USM','OVB'] or segments[0].origin in ['MOW','SHE'] and segments[0].destination in ['MOW','SHE'] or segments[0].origin in ['RTW','VRA'] and segments[0].destination in ['RTW','VRA'] or segments[0].origin in ['SHJ','VOZ'] and segments[0].destination in ['SHJ','VOZ'] or segments[0].origin in ['SSH','VOG'] and segments[0].destination in ['SSH','VOG'] or segments[0].origin in ['DXB','NOZ'] and segments[0].destination in ['DXB','NOZ'] or segments[0].origin in ['SGN','SGC'] and segments[0].destination in ['SGN','SGC'] or segments[0].origin in ['VVO','NHA'] and segments[0].destination in ['VVO','NHA'] or segments[0].origin in ['CUN','KZN'] and segments[0].destination in ['CUN','KZN'] or segments[0].origin in ['AYT','SVX'] and segments[0].destination in ['AYT','SVX'] or segments[0].origin in ['CUN','KGD'] and segments[0].destination in ['CUN','KGD'] or segments[0].origin in ['KBV','KZN'] and segments[0].destination in ['KBV','KZN'] or segments[0].origin in ['VRN','MOW'] and segments[0].destination in ['VRN','MOW'] or segments[0].origin in ['OVB','UUD'] and segments[0].destination in ['OVB','UUD'] or segments[0].origin in ['USM','TJM'] and segments[0].destination in ['USM','TJM'] or segments[0].origin in ['HRG','MMK'] and segments[0].destination in ['HRG','MMK'] or segments[0].origin in ['KUF','SSH'] and segments[0].destination in ['KUF','SSH'] or segments[0].origin in ['AER','LED'] and segments[0].destination in ['AER','LED'] or segments[0].origin in ['SGN','ROV'] and segments[0].destination in ['SGN','ROV'] or segments[0].origin in ['KZN','CUN'] and segments[0].destination in ['KZN','CUN'] or segments[0].origin in ['VRA','NBC'] and segments[0].destination in ['VRA','NBC'] or segments[0].origin in ['KUF','CUN'] and segments[0].destination in ['KUF','CUN'] or segments[0].origin in ['SSH','SGC'] and segments[0].destination in ['SSH','SGC'] or segments[0].origin in ['VRA','OVB'] and segments[0].destination in ['VRA','OVB'] or segments[0].origin in ['ODS','SKG'] and segments[0].destination in ['ODS','SKG'] or segments[0].origin in ['AMM','LED'] and segments[0].destination in ['AMM','LED'] or segments[0].origin in ['RTW','PUJ'] and segments[0].destination in ['RTW','PUJ'] or segments[0].origin in ['BKK','NJC'] and segments[0].destination in ['BKK','NJC'] or segments[0].origin in ['CUN','KRR'] and segments[0].destination in ['CUN','KRR'] or segments[0].origin in ['MRV','SSH'] and segments[0].destination in ['MRV','SSH'] or segments[0].origin in ['SGC','HRG'] and segments[0].destination in ['SGC','HRG'] or segments[0].origin in ['KZN','SKG'] and segments[0].destination in ['KZN','SKG'] or segments[0].origin in ['UFA','MOW'] and segments[0].destination in ['UFA','MOW'] or segments[0].origin in ['ROM','MOW'] and segments[0].destination in ['ROM','MOW'] or segments[0].origin in ['NBC','PUJ'] and segments[0].destination in ['NBC','PUJ'] or segments[0].origin in ['KHV','MOW'] and segments[0].destination in ['KHV','MOW'] or segments[0].origin in ['VRA','CEK'] and segments[0].destination in ['VRA','CEK'] or segments[0].origin in ['VRA','KEJ'] and segments[0].destination in ['VRA','KEJ'] or segments[0].origin in ['MOW','VVO'] and segments[0].destination in ['MOW','VVO'] or segments[0].origin in ['TOF','CUN'] and segments[0].destination in ['TOF','CUN'] or segments[0].origin in ['OVB','SKG'] and segments[0].destination in ['OVB','SKG'] or segments[0].origin in ['CUN','VOG'] and segments[0].destination in ['CUN','VOG'] or segments[0].origin in ['BKK','VOZ'] and segments[0].destination in ['BKK','VOZ'] or segments[0].origin in ['ROV','ETH'] and segments[0].destination in ['ROV','ETH'] or segments[0].origin in ['HTA','NHA'] and segments[0].destination in ['HTA','NHA'] or segments[0].origin in ['GOJ','VRA'] and segments[0].destination in ['GOJ','VRA'] or segments[0].origin in ['MOW','VRN'] and segments[0].destination in ['MOW','VRN'] or segments[0].origin in ['KZN','HRG'] and segments[0].destination in ['KZN','HRG'] or segments[0].origin in ['NHA','BAX'] and segments[0].destination in ['NHA','BAX'] or segments[0].origin in ['VRA','ASF'] and segments[0].destination in ['VRA','ASF'] or segments[0].origin in ['GOJ','SKG'] and segments[0].destination in ['GOJ','SKG'] or segments[0].origin in ['SKG','LWO'] and segments[0].destination in ['SKG','LWO'] or segments[0].origin in ['MRV','CUN'] and segments[0].destination in ['MRV','CUN'] or segments[0].origin in ['SOF','MOW'] and segments[0].destination in ['SOF','MOW'] or segments[0].origin in ['BAX','VRA'] and segments[0].destination in ['BAX','VRA'] or segments[0].origin in ['SSH','MRV'] and segments[0].destination in ['SSH','MRV'] or segments[0].origin in ['KRR','LED'] and segments[0].destination in ['KRR','LED'] or segments[0].origin in ['NHA','REN'] and segments[0].destination in ['NHA','REN'] or segments[0].origin in ['ATH','MOW'] and segments[0].destination in ['ATH','MOW'] or segments[0].origin in ['KZN','VRA'] and segments[0].destination in ['KZN','VRA'] or segments[0].origin in ['HRG','VOZ'] and segments[0].destination in ['HRG','VOZ'] or segments[0].origin in ['SGN','KUF'] and segments[0].destination in ['SGN','KUF'] or segments[0].origin in ['LED','CFU'] and segments[0].destination in ['LED','CFU'] or segments[0].origin in ['SGN','MRV'] and segments[0].destination in ['SGN','MRV'] or segments[0].origin in ['CUN','EGO'] and segments[0].destination in ['CUN','EGO'] or segments[0].origin in ['KJA','AER'] and segments[0].destination in ['KJA','AER'] or segments[0].origin in ['VRA','SCW'] and segments[0].destination in ['VRA','SCW'] or segments[0].origin in ['BQS','NHA'] and segments[0].destination in ['BQS','NHA'] or segments[0].origin in ['KGD','SSH'] and segments[0].destination in ['KGD','SSH'] or segments[0].origin in ['BKK','KRR'] and segments[0].destination in ['BKK','KRR'] or segments[0].origin in ['DXB','OVB'] and segments[0].destination in ['DXB','OVB'] or segments[0].origin in ['KRR','HRG'] and segments[0].destination in ['KRR','HRG'] or segments[0].origin in ['VRA','OMS'] and segments[0].destination in ['VRA','OMS'] or segments[0].origin in ['BKK','MRV'] and segments[0].destination in ['BKK','MRV'] or segments[0].origin in ['IKT','PUJ'] and segments[0].destination in ['IKT','PUJ'] or segments[0].origin in ['KZN','PUJ'] and segments[0].destination in ['KZN','PUJ'] or segments[0].origin in ['BKK','LED'] and segments[0].destination in ['BKK','LED'] or segments[0].origin in ['SGN','LED'] and segments[0].destination in ['SGN','LED'] or segments[0].origin in ['NHA','CEK'] and segments[0].destination in ['NHA','CEK'] or segments[0].origin in ['KJA','SSH'] and segments[0].destination in ['KJA','SSH'] or segments[0].origin in ['CUN','MOW'] and segments[0].destination in ['CUN','MOW'] or segments[0].origin in ['UUD','NHA'] and segments[0].destination in ['UUD','NHA'] or segments[0].origin in ['KUF','ETH'] and segments[0].destination in ['KUF','ETH'] or segments[0].origin in ['HKT','REN'] and segments[0].destination in ['HKT','REN'] or segments[0].origin in ['BKK','MOW'] and segments[0].destination in ['BKK','MOW'] or segments[0].origin in ['BKK','UUD'] and segments[0].destination in ['BKK','UUD'] or segments[0].origin in ['CUN','OVB'] and segments[0].destination in ['CUN','OVB'] or segments[0].origin in ['SVX','SSH'] and segments[0].destination in ['SVX','SSH'] or segments[0].origin in ['LED','ETH'] and segments[0].destination in ['LED','ETH'] or segments[0].origin in ['MSQ','CFU'] and segments[0].destination in ['MSQ','CFU'] or segments[0].origin in ['KGD','PUJ'] and segments[0].destination in ['KGD','PUJ'] or segments[0].origin in ['OVB','AER'] and segments[0].destination in ['OVB','AER'] or segments[0].origin in ['OMS','NHA'] and segments[0].destination in ['OMS','NHA'] or segments[0].origin in ['PUJ','GOJ'] and segments[0].destination in ['PUJ','GOJ'] or segments[0].origin in ['NHA','TOF'] and segments[0].destination in ['NHA','TOF'] or segments[0].origin in ['TDX','BAX'] and segments[0].destination in ['TDX','BAX'] or segments[0].origin in ['UTP','KJA'] and segments[0].destination in ['UTP','KJA'] or segments[0].origin in ['BKK','KHV'] and segments[0].destination in ['BKK','KHV'] or segments[0].origin in ['NHA','BQS'] and segments[0].destination in ['NHA','BQS'] or segments[0].origin in ['CMF','MOW'] and segments[0].destination in ['CMF','MOW'] or segments[0].origin in ['BER','MOW'] and segments[0].destination in ['BER','MOW'] or segments[0].origin in ['SGN','KHV'] and segments[0].destination in ['SGN','KHV'] or segments[0].origin in ['DXB','NJC'] and segments[0].destination in ['DXB','NJC'] or segments[0].origin in ['IKT','VRA'] and segments[0].destination in ['IKT','VRA'] or segments[0].origin in ['TAS','MOW'] and segments[0].destination in ['TAS','MOW'] or segments[0].origin in ['GOJ','AYT'] and segments[0].destination in ['GOJ','AYT'] or segments[0].origin in ['VRA','GOJ'] and segments[0].destination in ['VRA','GOJ'] or segments[0].origin in ['MOW','BQS'] and segments[0].destination in ['MOW','BQS'] or segments[0].origin in ['NOZ','VRA'] and segments[0].destination in ['NOZ','VRA'] or segments[0].origin in ['PUJ','CEK'] and segments[0].destination in ['PUJ','CEK'] or segments[0].origin in ['USM','BAX'] and segments[0].destination in ['USM','BAX'] or segments[0].origin in ['ROV','VRN'] and segments[0].destination in ['ROV','VRN'] or segments[0].origin in ['OVB','CUN'] and segments[0].destination in ['OVB','CUN'] or segments[0].origin in ['OVB','MOW'] and segments[0].destination in ['OVB','MOW'] or segments[0].origin in ['SKG','ROV'] and segments[0].destination in ['SKG','ROV'] or segments[0].origin in ['MOW','BKK'] and segments[0].destination in ['MOW','BKK'] or segments[0].origin in ['BKK','IKT'] and segments[0].destination in ['BKK','IKT'] or segments[0].origin in ['TDX','SGC'] and segments[0].destination in ['TDX','SGC'] or segments[0].origin in ['ROV','VRA'] and segments[0].destination in ['ROV','VRA'] or segments[0].origin in ['BKK','TOF'] and segments[0].destination in ['BKK','TOF'] or segments[0].origin in ['CUN','MRV'] and segments[0].destination in ['CUN','MRV'] or segments[0].origin in ['ZTH','MSQ'] and segments[0].destination in ['ZTH','MSQ'] or segments[0].origin in ['MOW','CMF'] and segments[0].destination in ['MOW','CMF'] or segments[0].origin in ['CUN','PEE'] and segments[0].destination in ['CUN','PEE'] or segments[0].origin in ['CEK','HRG'] and segments[0].destination in ['CEK','HRG'] or segments[0].origin in ['HRG','KRR'] and segments[0].destination in ['HRG','KRR'] or segments[0].origin in ['VAR','LED'] and segments[0].destination in ['VAR','LED'] or segments[0].origin in ['NBC','SSH'] and segments[0].destination in ['NBC','SSH'] or segments[0].origin in ['PUJ','AER'] and segments[0].destination in ['PUJ','AER'] or segments[0].origin in ['SIP','SVX'] and segments[0].destination in ['SIP','SVX'] or segments[0].origin in ['ROV','NHA'] and segments[0].destination in ['ROV','NHA'] or segments[0].origin in ['CUN','IKT'] and segments[0].destination in ['CUN','IKT'] or segments[0].origin in ['OVB','VRA'] and segments[0].destination in ['OVB','VRA'] or segments[0].origin in ['MOW','OVB'] and segments[0].destination in ['MOW','OVB'] or segments[0].origin in ['UUD','OVB'] and segments[0].destination in ['UUD','OVB'] or segments[0].origin in ['KRR','OVB'] and segments[0].destination in ['KRR','OVB'] or segments[0].origin in ['TJM','PUJ'] and segments[0].destination in ['TJM','PUJ'] or segments[0].origin in ['PEE','HRG'] and segments[0].destination in ['PEE','HRG'] or segments[0].origin in ['KZN','AYT'] and segments[0].destination in ['KZN','AYT'] or segments[0].origin in ['GVA','MOW'] and segments[0].destination in ['GVA','MOW'] or segments[0].origin in ['CUN','OGZ'] and segments[0].destination in ['CUN','OGZ'] or segments[0].origin in ['MUC','MOW'] and segments[0].destination in ['MUC','MOW'] or segments[0].origin in ['VOZ','SSH'] and segments[0].destination in ['VOZ','SSH'] or segments[0].origin in ['AER','OVB'] and segments[0].destination in ['AER','OVB'] or segments[0].origin in ['HRG','KEJ'] and segments[0].destination in ['HRG','KEJ'] or segments[0].origin in ['TJM','VRA'] and segments[0].destination in ['TJM','VRA'] or segments[0].origin in ['HKT','BAX'] and segments[0].destination in ['HKT','BAX'] or segments[0].origin in ['KUF','AER'] and segments[0].destination in ['KUF','AER'] or segments[0].origin in ['SGN','HTA'] and segments[0].destination in ['SGN','HTA'] or segments[0].origin in ['SSH','UFA'] and segments[0].destination in ['SSH','UFA'] or segments[0].origin in ['SHJ','MOW'] and segments[0].destination in ['SHJ','MOW'] or segments[0].origin in ['SSH','KZN'] and segments[0].destination in ['SSH','KZN'] or segments[0].origin in ['SVX','PUJ'] and segments[0].destination in ['SVX','PUJ'] or segments[0].origin in ['PRG','MOW'] and segments[0].destination in ['PRG','MOW'] or segments[0].origin in ['VOZ','VRA'] and segments[0].destination in ['VOZ','VRA'] or segments[0].origin in ['AER','MOW'] and segments[0].destination in ['AER','MOW'] or segments[0].origin in ['SSH','OMS'] and segments[0].destination in ['SSH','OMS'] or segments[0].origin in ['SSH','SCW'] and segments[0].destination in ['SSH','SCW'] or segments[0].origin in ['CUN','MCX'] and segments[0].destination in ['CUN','MCX'] or segments[0].origin in ['MMK','HRG'] and segments[0].destination in ['MMK','HRG'] or segments[0].origin in ['LED','SOF'] and segments[0].destination in ['LED','SOF'] or segments[0].origin in ['KBV','UFA'] and segments[0].destination in ['KBV','UFA'] or segments[0].origin in ['DJE','MOW'] and segments[0].destination in ['DJE','MOW'] or segments[0].origin in ['NJC','VRA'] and segments[0].destination in ['NJC','VRA'] or segments[0].origin in ['YKS','NHA'] and segments[0].destination in ['YKS','NHA'] or segments[0].origin in ['SSH','MMK'] and segments[0].destination in ['SSH','MMK'] or segments[0].origin in ['PUJ','TJM'] and segments[0].destination in ['PUJ','TJM'] or segments[0].origin in ['TOF','NHA'] and segments[0].destination in ['TOF','NHA'] or segments[0].origin in ['SGN','PEE'] and segments[0].destination in ['SGN','PEE'] or segments[0].origin in ['NOZ','CUN'] and segments[0].destination in ['NOZ','CUN'] or segments[0].origin in ['PEE','PUJ'] and segments[0].destination in ['PEE','PUJ'] or segments[0].origin in ['SVX','NHA'] and segments[0].destination in ['SVX','NHA'] or segments[0].origin in ['ARH','NHA'] and segments[0].destination in ['ARH','NHA'] or segments[0].origin in ['SCW','NHA'] and segments[0].destination in ['SCW','NHA'] or segments[0].origin in ['KEJ','SSH'] and segments[0].destination in ['KEJ','SSH'] or segments[0].origin in ['AER','UFA'] and segments[0].destination in ['AER','UFA'] or segments[0].origin in ['NHA','MCX'] and segments[0].destination in ['NHA','MCX'] or segments[0].origin in ['CUN','LED'] and segments[0].destination in ['CUN','LED'] or segments[0].origin in ['MOW','FEG'] and segments[0].destination in ['MOW','FEG'] or segments[0].origin in ['MOW','SVX'] and segments[0].destination in ['MOW','SVX'] or segments[0].origin in ['KBV','SGC'] and segments[0].destination in ['KBV','SGC'] or segments[0].origin in ['VRA','KRR'] and segments[0].destination in ['VRA','KRR'] or segments[0].origin in ['SKG','KRR'] and segments[0].destination in ['SKG','KRR'] or segments[0].origin in ['NJC','PUJ'] and segments[0].destination in ['NJC','PUJ'] or segments[0].origin in ['MSQ','ZTH'] and segments[0].destination in ['MSQ','ZTH'] or segments[0].origin in ['SKG','VOG'] and segments[0].destination in ['SKG','VOG'] or segments[0].origin in ['KJA','CUN'] and segments[0].destination in ['KJA','CUN'] or segments[0].origin in ['DXB','GOJ'] and segments[0].destination in ['DXB','GOJ'] or segments[0].origin in ['SGN','BAX'] and segments[0].destination in ['SGN','BAX'] or segments[0].origin in ['KUF','AYT'] and segments[0].destination in ['KUF','AYT'] or segments[0].origin in ['ETH','KRR'] and segments[0].destination in ['ETH','KRR'] or segments[0].origin in ['IKT','NHA'] and segments[0].destination in ['IKT','NHA'] or segments[0].origin in ['ROV','HRG'] and segments[0].destination in ['ROV','HRG'] or segments[0].origin in ['PUJ','IKT'] and segments[0].destination in ['PUJ','IKT'] or segments[0].origin in ['TIV','MOW'] and segments[0].destination in ['TIV','MOW'] or segments[0].origin in ['PUJ','MOW'] and segments[0].destination in ['PUJ','MOW'] or segments[0].origin in ['CEK','VRA'] and segments[0].destination in ['CEK','VRA'] or segments[0].origin in ['EGO','PUJ'] and segments[0].destination in ['EGO','PUJ'] or segments[0].origin in ['TDX','IKT'] and segments[0].destination in ['TDX','IKT'] or segments[0].origin in ['SKG','KGD'] and segments[0].destination in ['SKG','KGD'] or segments[0].origin in ['SGN','UFA'] and segments[0].destination in ['SGN','UFA'] or segments[0].origin in ['MOW','BOJ'] and segments[0].destination in ['MOW','BOJ'] or segments[0].origin in ['NHA','KRR'] and segments[0].destination in ['NHA','KRR'] or segments[0].origin in ['HKT','KHV'] and segments[0].destination in ['HKT','KHV'] or segments[0].origin in ['RIX','SKG'] and segments[0].destination in ['RIX','SKG'] or segments[0].origin in ['SIP','KRR'] and segments[0].destination in ['SIP','KRR'] or segments[0].origin in ['AAQ','VRA'] and segments[0].destination in ['AAQ','VRA'] or segments[0].origin in ['VOZ','HRG'] and segments[0].destination in ['VOZ','HRG'] or segments[0].origin in ['CFU','LED'] and segments[0].destination in ['CFU','LED'] or segments[0].origin in ['KBV','BQS'] and segments[0].destination in ['KBV','BQS'] or segments[0].origin in ['BKK','NBC'] and segments[0].destination in ['BKK','NBC'] or segments[0].origin in ['SSH','GOJ'] and segments[0].destination in ['SSH','GOJ'] or segments[0].origin in ['LED','OVB'] and segments[0].destination in ['LED','OVB'] or segments[0].origin in ['NHA','UUD'] and segments[0].destination in ['NHA','UUD'] or segments[0].origin in ['CUN','UFA'] and segments[0].destination in ['CUN','UFA'] or segments[0].origin in ['MMK','SSH'] and segments[0].destination in ['MMK','SSH'] or segments[0].origin in ['MOW','PKC'] and segments[0].destination in ['MOW','PKC'] or segments[0].origin in ['SKG','ODS'] and segments[0].destination in ['SKG','ODS'] or segments[0].origin in ['UFA','SKG'] and segments[0].destination in ['UFA','SKG'] or segments[0].origin in ['UFA','AER'] and segments[0].destination in ['UFA','AER'] or segments[0].origin in ['VRA','NOZ'] and segments[0].destination in ['VRA','NOZ'] or segments[0].origin in ['NHA','MOW'] and segments[0].destination in ['NHA','MOW'] or segments[0].origin in ['HKT','NOZ'] and segments[0].destination in ['HKT','NOZ'] or segments[0].origin in ['MCX','VRA'] and segments[0].destination in ['MCX','VRA'] or segments[0].origin in ['SIP','LED'] and segments[0].destination in ['SIP','LED'] or segments[0].origin in ['MOW','BGY'] and segments[0].destination in ['MOW','BGY'] or segments[0].origin in ['HKT','EGO'] and segments[0].destination in ['HKT','EGO'] or segments[0].origin in ['KZN','AER'] and segments[0].destination in ['KZN','AER'] or segments[0].origin in ['NHA','OVB'] and segments[0].destination in ['NHA','OVB'] or segments[0].origin in ['VRA','VOZ'] and segments[0].destination in ['VRA','VOZ'] or segments[0].origin in ['OVB','LED'] and segments[0].destination in ['OVB','LED'] or segments[0].origin in ['NBC','CUN'] and segments[0].destination in ['NBC','CUN'] or segments[0].origin in ['VRA','KGD'] and segments[0].destination in ['VRA','KGD'] or segments[0].origin in ['CUN','CEK'] and segments[0].destination in ['CUN','CEK'] or segments[0].origin in ['VOZ','CUN'] and segments[0].destination in ['VOZ','CUN'] or segments[0].origin in ['DYR','MOW'] and segments[0].destination in ['DYR','MOW'] or segments[0].origin in ['MOW','SOF'] and segments[0].destination in ['MOW','SOF'] or segments[0].origin in ['LED','PRG'] and segments[0].destination in ['LED','PRG'] or segments[0].origin in ['PKC','NHA'] and segments[0].destination in ['PKC','NHA'] or segments[0].origin in ['BKK','TJM'] and segments[0].destination in ['BKK','TJM'] or segments[0].origin in ['NHA','OMS'] and segments[0].destination in ['NHA','OMS'] or segments[0].origin in ['DXB','BAX'] and segments[0].destination in ['DXB','BAX'] or segments[0].origin in ['OVB','HRG'] and segments[0].destination in ['OVB','HRG'] or segments[0].origin in ['AYT','KUF'] and segments[0].destination in ['AYT','KUF'] or segments[0].origin in ['HKT','CEK'] and segments[0].destination in ['HKT','CEK'] or segments[0].origin in ['GRV','MOW'] and segments[0].destination in ['GRV','MOW'] or segments[0].origin in ['IEV','ATH'] and segments[0].destination in ['IEV','ATH'] or segments[0].origin in ['OGZ','NHA'] and segments[0].destination in ['OGZ','NHA'] or segments[0].origin in ['ROV','SSH'] and segments[0].destination in ['ROV','SSH'] or segments[0].origin in ['SKG','UFA'] and segments[0].destination in ['SKG','UFA'] or segments[0].origin in ['CUN','BAX'] and segments[0].destination in ['CUN','BAX'] or segments[0].origin in ['SZG','MOW'] and segments[0].destination in ['SZG','MOW'] or segments[0].origin in ['HKT','KGD'] and segments[0].destination in ['HKT','KGD'] or segments[0].origin in ['ROV','SKG'] and segments[0].destination in ['ROV','SKG'] or segments[0].origin in ['USM','SVX'] and segments[0].destination in ['USM','SVX'] or segments[0].origin in ['KBV','BAX'] and segments[0].destination in ['KBV','BAX'] or segments[0].origin in ['BQS','MOW'] and segments[0].destination in ['BQS','MOW'] or segments[0].origin in ['SSH','KEJ'] and segments[0].destination in ['SSH','KEJ'] or segments[0].origin in ['SIP','UFA'] and segments[0].destination in ['SIP','UFA'] or segments[0].origin in ['CUN','YKS'] and segments[0].destination in ['CUN','YKS'] or segments[0].origin in ['GOJ','NHA'] and segments[0].destination in ['GOJ','NHA'] or segments[0].origin in ['MOW','PUJ'] and segments[0].destination in ['MOW','PUJ'] or segments[0].origin in ['NHA','LED'] and segments[0].destination in ['NHA','LED'] or segments[0].origin in ['HKT','VOZ'] and segments[0].destination in ['HKT','VOZ'] or segments[0].origin in ['OMS','VRA'] and segments[0].destination in ['OMS','VRA'] or segments[0].origin in ['OVB','BQS'] and segments[0].destination in ['OVB','BQS'] or segments[0].origin in ['BKK','GOJ'] and segments[0].destination in ['BKK','GOJ'] or segments[0].origin in ['HKT','ASF'] and segments[0].destination in ['HKT','ASF'] or segments[0].origin in ['LED','PUJ'] and segments[0].destination in ['LED','PUJ'] or segments[0].origin in ['CUN','KUF'] and segments[0].destination in ['CUN','KUF'] or segments[0].origin in ['MOW','LCA'] and segments[0].destination in ['MOW','LCA'] or segments[0].origin in ['CUN','KEJ'] and segments[0].destination in ['CUN','KEJ'] or segments[0].origin in ['LWO','SKG'] and segments[0].destination in ['LWO','SKG'] or segments[0].origin in ['HRG','SVX'] and segments[0].destination in ['HRG','SVX'] or segments[0].origin in ['TCI','MOW'] and segments[0].destination in ['TCI','MOW'] or segments[0].origin in ['SIP','AER'] and segments[0].destination in ['SIP','AER'] or segments[0].origin in ['SGN','TJM'] and segments[0].destination in ['SGN','TJM'] or segments[0].origin in ['PUJ','VOG'] and segments[0].destination in ['PUJ','VOG'] or segments[0].origin in ['UFA','SSH'] and segments[0].destination in ['UFA','SSH'] or segments[0].origin in ['MIL','MOW'] and segments[0].destination in ['MIL','MOW'] or segments[0].origin in ['AER','PUJ'] and segments[0].destination in ['AER','PUJ'] or segments[0].origin in ['NHA','HTA'] and segments[0].destination in ['NHA','HTA'] or segments[0].origin in ['BQS','OVB'] and segments[0].destination in ['BQS','OVB'] or segments[0].origin in ['USM','MOW'] and segments[0].destination in ['USM','MOW'] or segments[0].origin in ['KBV','IKT'] and segments[0].destination in ['KBV','IKT'] or segments[0].origin in ['HKT','UFA'] and segments[0].destination in ['HKT','UFA'] or segments[0].origin in ['MOW','KHV'] and segments[0].destination in ['MOW','KHV'] or segments[0].origin in ['UTP','EGO'] and segments[0].destination in ['UTP','EGO'] or segments[0].origin in ['DXB','HTA'] and segments[0].destination in ['DXB','HTA'] or segments[0].origin in ['SGN','OMS'] and segments[0].destination in ['SGN','OMS'] or segments[0].origin in ['MOW','AER'] and segments[0].destination in ['MOW','AER'] or segments[0].origin in ['HTA','PUJ'] and segments[0].destination in ['HTA','PUJ'] or segments[0].origin in ['KJA','NHA'] and segments[0].destination in ['KJA','NHA'] or segments[0].origin in ['HKT','OMS'] and segments[0].destination in ['HKT','OMS'] or segments[0].origin in ['OGZ','PUJ'] and segments[0].destination in ['OGZ','PUJ'] or segments[0].origin in ['PUJ','UFA'] and segments[0].destination in ['PUJ','UFA'] or segments[0].origin in ['DXB','KUF'] and segments[0].destination in ['DXB','KUF'] or segments[0].origin in ['BKK','MCX'] and segments[0].destination in ['BKK','MCX'] or segments[0].origin in ['NHA','PKC'] and segments[0].destination in ['NHA','PKC'] or segments[0].origin in ['CUN','KJA'] and segments[0].destination in ['CUN','KJA'] or segments[0].origin in ['KRR','PUJ'] and segments[0].destination in ['KRR','PUJ'] or segments[0].origin in ['HKT','IKT'] and segments[0].destination in ['HKT','IKT'] or segments[0].origin in ['DXB','ROV'] and segments[0].destination in ['DXB','ROV'] or segments[0].origin in ['DXB','TJM'] and segments[0].destination in ['DXB','TJM'] or segments[0].origin in ['NHA','KJA'] and segments[0].destination in ['NHA','KJA'] or segments[0].origin in ['USM','OMS'] and segments[0].destination in ['USM','OMS'] or segments[0].origin in ['KHV','NHA'] and segments[0].destination in ['KHV','NHA'] or segments[0].origin in ['HRG','KGD'] and segments[0].destination in ['HRG','KGD'] or segments[0].origin in ['VOG','SSH'] and segments[0].destination in ['VOG','SSH'] or segments[0].origin in ['MCX','PUJ'] and segments[0].destination in ['MCX','PUJ'] or segments[0].origin in ['MOW','TIV'] and segments[0].destination in ['MOW','TIV'] or segments[0].origin in ['DXB','KRR'] and segments[0].destination in ['DXB','KRR'] or segments[0].origin in ['DNK','SKG'] and segments[0].destination in ['DNK','SKG'] or segments[0].origin in ['HKT','KZN'] and segments[0].destination in ['HKT','KZN'] or segments[0].origin in ['USM','LED'] and segments[0].destination in ['USM','LED'] or segments[0].origin in ['HKT','MRV'] and segments[0].destination in ['HKT','MRV'] or segments[0].origin in ['HKT','TOF'] and segments[0].destination in ['HKT','TOF'] or segments[0].origin in ['MOW','UFA'] and segments[0].destination in ['MOW','UFA'] or segments[0].origin in ['DXB','KEJ'] and segments[0].destination in ['DXB','KEJ'] or segments[0].origin in ['YKS','CUN'] and segments[0].destination in ['YKS','CUN'] or segments[0].origin in ['KEJ','HRG'] and segments[0].destination in ['KEJ','HRG'] or segments[0].origin in ['MCX','NHA'] and segments[0].destination in ['MCX','NHA'] or segments[0].origin in ['NHA','SCW'] and segments[0].destination in ['NHA','SCW'] or segments[0].origin in ['DXB','MRV'] and segments[0].destination in ['DXB','MRV'] or segments[0].origin in ['BKK','OGZ'] and segments[0].destination in ['BKK','OGZ'] or segments[0].origin in ['UTP','PEE'] and segments[0].destination in ['UTP','PEE'] or segments[0].origin in ['USM','ROV'] and segments[0].destination in ['USM','ROV'] or segments[0].origin in ['VRA','YKS'] and segments[0].destination in ['VRA','YKS'] or segments[0].origin in ['SHE','MOW'] and segments[0].destination in ['SHE','MOW'] or segments[0].origin in ['MOW','TSN'] and segments[0].destination in ['MOW','TSN'] or segments[0].origin in ['TOF','OVB'] and segments[0].destination in ['TOF','OVB'] or segments[0].origin in ['NHA','KEJ'] and segments[0].destination in ['NHA','KEJ'] or segments[0].origin in ['KGD','CUN'] and segments[0].destination in ['KGD','CUN'] or segments[0].origin in ['UTP','KUF'] and segments[0].destination in ['UTP','KUF'] or segments[0].origin in ['SIP','KZN'] and segments[0].destination in ['SIP','KZN'] or segments[0].origin in ['CUN','SCW'] and segments[0].destination in ['CUN','SCW'] or segments[0].origin in ['SHJ','REN'] and segments[0].destination in ['SHJ','REN'] or segments[0].origin in ['SGN','KRR'] and segments[0].destination in ['SGN','KRR'] or segments[0].origin in ['KEJ','NHA'] and segments[0].destination in ['KEJ','NHA'] or segments[0].origin in ['CFU','IEV'] and segments[0].destination in ['CFU','IEV'] or segments[0].origin in ['MOW','CUN'] and segments[0].destination in ['MOW','CUN'] or segments[0].origin in ['LCA','MOW'] and segments[0].destination in ['LCA','MOW'] or segments[0].origin in ['SSH','ROV'] and segments[0].destination in ['SSH','ROV'] or segments[0].origin in ['BUH','MOW'] and segments[0].destination in ['BUH','MOW'] or segments[0].origin in ['SGN','BQS'] and segments[0].destination in ['SGN','BQS'] or segments[0].origin in ['KUF','VRA'] and segments[0].destination in ['KUF','VRA'] or segments[0].origin in ['NHA','KHV'] and segments[0].destination in ['NHA','KHV'] or segments[0].origin in ['DXB','TOF'] and segments[0].destination in ['DXB','TOF'] or segments[0].origin in ['HKT','KUF'] and segments[0].destination in ['HKT','KUF'] or segments[0].origin in ['EGO','NHA'] and segments[0].destination in ['EGO','NHA'] or segments[0].origin in ['MOW','BCN'] and segments[0].destination in ['MOW','BCN'] or segments[0].origin in ['SCW','HRG'] and segments[0].destination in ['SCW','HRG'] or segments[0].origin in ['BAX','CUN'] and segments[0].destination in ['BAX','CUN'] or segments[0].origin in ['AYT','PEE'] and segments[0].destination in ['AYT','PEE'] or segments[0].origin in ['BKK','OMS'] and segments[0].destination in ['BKK','OMS'] or segments[0].origin in ['LCA','KRR'] and segments[0].destination in ['LCA','KRR'] or segments[0].origin in ['BKK','CEK'] and segments[0].destination in ['BKK','CEK'] or segments[0].origin in ['MOW','VRA'] and segments[0].destination in ['MOW','VRA'] or segments[0].origin in ['LED','ZTH'] and segments[0].destination in ['LED','ZTH'] or segments[0].origin in ['KEJ','VRA'] and segments[0].destination in ['KEJ','VRA'] or segments[0].origin in ['MOW','DYR'] and segments[0].destination in ['MOW','DYR'] or segments[0].origin in ['HKT','YKS'] and segments[0].destination in ['HKT','YKS'] or segments[0].origin in ['MOW','MIR'] and segments[0].destination in ['MOW','MIR'] or segments[0].origin in ['TRN','MOW'] and segments[0].destination in ['TRN','MOW'] or segments[0].origin in ['RVN','MOW'] and segments[0].destination in ['RVN','MOW'] or segments[0].origin in ['CEK','SSH'] and segments[0].destination in ['CEK','SSH'] or segments[0].origin in ['ETH','UFA'] and segments[0].destination in ['ETH','UFA'] or segments[0].origin in ['VRA','UFA'] and segments[0].destination in ['VRA','UFA'] or segments[0].origin in ['MOW','HER'] and segments[0].destination in ['MOW','HER'] or segments[0].origin in ['DXB','OMS'] and segments[0].destination in ['DXB','OMS'] or segments[0].origin in ['VRA','ROV'] and segments[0].destination in ['VRA','ROV'] or segments[0].origin in ['MRV','PUJ'] and segments[0].destination in ['MRV','PUJ'] or segments[0].origin in ['NHA','EGO'] and segments[0].destination in ['NHA','EGO'] or segments[0].origin in ['VRA','TOF'] and segments[0].destination in ['VRA','TOF'] or segments[0].origin in ['BOJ','LED'] and segments[0].destination in ['BOJ','LED'] or segments[0].origin in ['MOW','BHK'] and segments[0].destination in ['MOW','BHK'] or segments[0].origin in ['HKT','VVO'] and segments[0].destination in ['HKT','VVO'] or segments[0].origin in ['TOF','MOW'] and segments[0].destination in ['TOF','MOW'] or segments[0].origin in ['USM','KZN'] and segments[0].destination in ['USM','KZN'] or segments[0].origin in ['PUJ','KUF'] and segments[0].destination in ['PUJ','KUF'] or segments[0].origin in ['VOZ','PUJ'] and segments[0].destination in ['VOZ','PUJ'] or segments[0].origin in ['OVB','KRR'] and segments[0].destination in ['OVB','KRR'] or segments[0].origin in ['MOW','IKT'] and segments[0].destination in ['MOW','IKT'] or segments[0].origin in ['PEE','VRA'] and segments[0].destination in ['PEE','VRA'] or segments[0].origin in ['CFU','ROV'] and segments[0].destination in ['CFU','ROV'] or segments[0].origin in ['POP','MOW'] and segments[0].destination in ['POP','MOW'] or segments[0].origin in ['PUJ','SCW'] and segments[0].destination in ['PUJ','SCW'] or segments[0].origin in ['BAX','MOW'] and segments[0].destination in ['BAX','MOW'] or segments[0].origin in ['PUJ','SVX'] and segments[0].destination in ['PUJ','SVX'] or segments[0].origin in ['CUN','NJC'] and segments[0].destination in ['CUN','NJC'] or segments[0].origin in ['UTP','LED'] and segments[0].destination in ['UTP','LED'] or segments[0].origin in ['NHA','TJM'] and segments[0].destination in ['NHA','TJM'] or segments[0].origin in ['SGN','GOJ'] and segments[0].destination in ['SGN','GOJ'] or segments[0].origin in ['SSH','NBC'] and segments[0].destination in ['SSH','NBC'] or segments[0].origin in ['KJA','MOW'] and segments[0].destination in ['KJA','MOW'] or segments[0].origin in ['MOW','GPA'] and segments[0].destination in ['MOW','GPA'] or segments[0].origin in ['ATH','IEV'] and segments[0].destination in ['ATH','IEV'] or segments[0].origin in ['USM','VVO'] and segments[0].destination in ['USM','VVO'] or segments[0].origin in ['MOW','RMI'] and segments[0].destination in ['MOW','RMI'] or segments[0].origin in ['CEE','PUJ'] and segments[0].destination in ['CEE','PUJ'] or segments[0].origin in ['KRR','SKG'] and segments[0].destination in ['KRR','SKG'] or segments[0].origin in ['CUN','HTA'] and segments[0].destination in ['CUN','HTA'] or segments[0].origin in ['MRV','VRA'] and segments[0].destination in ['MRV','VRA'] or segments[0].origin in ['VRA','TJM'] and segments[0].destination in ['VRA','TJM'] or segments[0].origin in ['SKG','RIX'] and segments[0].destination in ['SKG','RIX'] or segments[0].origin in ['PRG','SVX'] and segments[0].destination in ['PRG','SVX'] or segments[0].origin in ['ABA','VRA'] and segments[0].destination in ['ABA','VRA'] or segments[0].origin in ['SGN','IKT'] and segments[0].destination in ['SGN','IKT'] or segments[0].origin in ['VOG','HRG'] and segments[0].destination in ['VOG','HRG'] or segments[0].origin in ['SVX','HER'] and segments[0].destination in ['SVX','HER'] or segments[0].origin in ['SHJ','VOG'] and segments[0].destination in ['SHJ','VOG'] or segments[0].origin in ['VRA','OGZ'] and segments[0].destination in ['VRA','OGZ'] or segments[0].origin in ['MOW','ZTH'] and segments[0].destination in ['MOW','ZTH'] or segments[0].origin in ['KJA','PUJ'] and segments[0].destination in ['KJA','PUJ'] or segments[0].origin in ['SSH','KJA'] and segments[0].destination in ['SSH','KJA'] or segments[0].origin in ['PUJ','NBC'] and segments[0].destination in ['PUJ','NBC'] or segments[0].origin in ['BKK','BAX'] and segments[0].destination in ['BKK','BAX'] or segments[0].origin in ['GOJ','HKT'] and segments[0].destination in ['GOJ','HKT'] or segments[0].origin in ['LED','AYT'] and segments[0].destination in ['LED','AYT'] or segments[0].origin in ['CEK','USM'] and segments[0].destination in ['CEK','USM'] or segments[0].origin in ['LED','SHJ'] and segments[0].destination in ['LED','SHJ'] or segments[0].origin in ['NOZ','BKK'] and segments[0].destination in ['NOZ','BKK'] or segments[0].origin in ['NOZ','PUJ'] and segments[0].destination in ['NOZ','PUJ'] or segments[0].origin in ['TJM','TDX'] and segments[0].destination in ['TJM','TDX'] or segments[0].origin in ['YKS','BKK'] and segments[0].destination in ['YKS','BKK'] or segments[0].origin in ['MOW','KUF'] and segments[0].destination in ['MOW','KUF'] or segments[0].origin in ['KJA','SGN'] and segments[0].destination in ['KJA','SGN'] or segments[0].origin in ['CEK','UTP'] and segments[0].destination in ['CEK','UTP'] or segments[0].origin in ['UFA','USM'] and segments[0].destination in ['UFA','USM'] or segments[0].origin in ['KZN','TDX'] and segments[0].destination in ['KZN','TDX'] or segments[0].origin in ['PEE','DXB'] and segments[0].destination in ['PEE','DXB'] or segments[0].origin in ['NJC','HKT'] and segments[0].destination in ['NJC','HKT'] or segments[0].origin in ['UFA','BKK'] and segments[0].destination in ['UFA','BKK'] or segments[0].origin in ['VOG','BKK'] and segments[0].destination in ['VOG','BKK'] or segments[0].origin in ['CEK','SGN'] and segments[0].destination in ['CEK','SGN'] or segments[0].origin in ['KZN','UTP'] and segments[0].destination in ['KZN','UTP'] or segments[0].origin in ['KJA','HKT'] and segments[0].destination in ['KJA','HKT'] or segments[0].origin in ['HTA','BKK'] and segments[0].destination in ['HTA','BKK'] or segments[0].origin in ['PEE','TDX'] and segments[0].destination in ['PEE','TDX'] or segments[0].origin in ['OVB','SGN'] and segments[0].destination in ['OVB','SGN'] or segments[0].origin in ['AER','HKT'] and segments[0].destination in ['AER','HKT'] or segments[0].origin in ['CUN','CEE'] and segments[0].destination in ['CUN','CEE'] or segments[0].origin in ['OGZ','MOW'] and segments[0].destination in ['OGZ','MOW'] or segments[0].origin in ['SVX','AQJ'] and segments[0].destination in ['SVX','AQJ'] or segments[0].origin in ['SVX','SHJ'] and segments[0].destination in ['SVX','SHJ'] or segments[0].origin in ['NHA','BTK'] and segments[0].destination in ['NHA','BTK'] or segments[0].origin in ['KZN','AQJ'] and segments[0].destination in ['KZN','AQJ'] or segments[0].origin in ['VOZ','DXB'] and segments[0].destination in ['VOZ','DXB'] or segments[0].origin in ['OGZ','HKT'] and segments[0].destination in ['OGZ','HKT'] or segments[0].origin in ['MOW','KLV'] and segments[0].destination in ['MOW','KLV'] or segments[0].origin in ['MOW','AQJ'] and segments[0].destination in ['MOW','AQJ'] or segments[0].origin in ['AER','SGN'] and segments[0].destination in ['AER','SGN'] or segments[0].origin in ['OVB','BKK'] and segments[0].destination in ['OVB','BKK'] or segments[0].origin in ['UFA','AYT'] and segments[0].destination in ['UFA','AYT'] or segments[0].origin in ['NOZ','SGN'] and segments[0].destination in ['NOZ','SGN'] or segments[0].origin in ['NBC','SGN'] and segments[0].destination in ['NBC','SGN'] or segments[0].origin in ['BQS','TDX'] and segments[0].destination in ['BQS','TDX'] or segments[0].origin in ['SGC','NHA'] and segments[0].destination in ['SGC','NHA'] or segments[0].origin in ['UFA','NHA'] and segments[0].destination in ['UFA','NHA'] or segments[0].origin in ['ROV','TDX'] and segments[0].destination in ['ROV','TDX'] or segments[0].origin in ['KZN','BKK'] and segments[0].destination in ['KZN','BKK'] or segments[0].origin in ['KJA','DXB'] and segments[0].destination in ['KJA','DXB'] or segments[0].origin in ['CEK','DXB'] and segments[0].destination in ['CEK','DXB'] or segments[0].origin in ['OVB','KBV'] and segments[0].destination in ['OVB','KBV'] or segments[0].origin in ['PEE','KBV'] and segments[0].destination in ['PEE','KBV'] or segments[0].origin in ['SVX','LED'] and segments[0].destination in ['SVX','LED'] or segments[0].origin in ['CUN','AAQ'] and segments[0].destination in ['CUN','AAQ'] or segments[0].origin in ['KEJ','BKK'] and segments[0].destination in ['KEJ','BKK'] or segments[0].origin in ['BQS','BKK'] and segments[0].destination in ['BQS','BKK'] or segments[0].origin in ['IKT','DXB'] and segments[0].destination in ['IKT','DXB'] or segments[0].origin in ['KUF','BKK'] and segments[0].destination in ['KUF','BKK'] or segments[0].origin in ['VVO','BKK'] and segments[0].destination in ['VVO','BKK'] or segments[0].origin in ['OVB','HKT'] and segments[0].destination in ['OVB','HKT'] or segments[0].origin in ['NHA','RTW'] and segments[0].destination in ['NHA','RTW'] or segments[0].origin in ['VOG','SGN'] and segments[0].destination in ['VOG','SGN'] or segments[0].origin in ['VVO','KBV'] and segments[0].destination in ['VVO','KBV'] or segments[0].origin in ['TOF','PUJ'] and segments[0].destination in ['TOF','PUJ'] or segments[0].origin in ['KEJ','HKT'] and segments[0].destination in ['KEJ','HKT'] or segments[0].origin in ['TJM','HKT'] and segments[0].destination in ['TJM','HKT'] or segments[0].origin in ['ROV','POP'] and segments[0].destination in ['ROV','POP'] or segments[0].origin in ['REN','ETH'] and segments[0].destination in ['REN','ETH'] or segments[0].origin in ['PEE','USM'] and segments[0].destination in ['PEE','USM'] or segments[0].origin in ['KEJ','SGN'] and segments[0].destination in ['KEJ','SGN'] or segments[0].origin in ['AER','DXB'] and segments[0].destination in ['AER','DXB'] or segments[0].origin in ['VOG','MOW'] and segments[0].destination in ['VOG','MOW'] or segments[0].origin in ['YKS','SGN'] and segments[0].destination in ['YKS','SGN'] or segments[0].origin in ['UFA','TDX'] and segments[0].destination in ['UFA','TDX'] or segments[0].origin in ['CUN','ABA'] and segments[0].destination in ['CUN','ABA'] or segments[0].origin in ['HRG','TJM'] and segments[0].destination in ['HRG','TJM'] or segments[0].origin in ['HTA','HKT'] and segments[0].destination in ['HTA','HKT'] or segments[0].origin in ['NBC','HKT'] and segments[0].destination in ['NBC','HKT'] or segments[0].origin in ['TOF','SGN'] and segments[0].destination in ['TOF','SGN'] or segments[0].origin in ['MCX','HKT'] and segments[0].destination in ['MCX','HKT'] or segments[0].origin in ['VRA','CEE'] and segments[0].destination in ['VRA','CEE'] or segments[0].origin in ['CUN','RTW'] and segments[0].destination in ['CUN','RTW'] or segments[0].origin in ['VOZ','SGN'] and segments[0].destination in ['VOZ','SGN'] or segments[0].origin in ['BQS','USM'] and segments[0].destination in ['BQS','USM'] or segments[0].origin in ['SGC','USM'] and segments[0].destination in ['SGC','USM'] or segments[0].origin in ['KZN','DXB'] and segments[0].destination in ['KZN','DXB'] or segments[0].origin in ['KEJ','TDX'] and segments[0].destination in ['KEJ','TDX'] or segments[0].origin in ['UFA','DXB'] and segments[0].destination in ['UFA','DXB'] or segments[0].origin in ['LED','HKT'] and segments[0].destination in ['LED','HKT'] or segments[0].origin in ['NJC','SGN'] and segments[0].destination in ['NJC','SGN'] or segments[0].origin in ['KUF','SHJ'] and segments[0].destination in ['KUF','SHJ'] or segments[0].origin in ['LED','AQJ'] and segments[0].destination in ['LED','AQJ'] or segments[0].origin in ['KJA','SIP'] and segments[0].destination in ['KJA','SIP'] or segments[0].origin in ['SGC','BKK'] and segments[0].destination in ['SGC','BKK'] or segments[0].origin in ['KJA','BKK'] and segments[0].destination in ['KJA','BKK'] or segments[0].origin in ['VOG','DXB'] and segments[0].destination in ['VOG','DXB'] or segments[0].origin in ['KEJ','USM'] and segments[0].destination in ['KEJ','USM'] or segments[0].origin in ['VVO','SGN'] and segments[0].destination in ['VVO','SGN'] or segments[0].origin in ['MOW','SIP'] and segments[0].destination in ['MOW','SIP'] or segments[0].origin in ['MOW','ROV'] and segments[0].destination in ['MOW','ROV'] or segments[0].origin in ['TJM','KBV'] and segments[0].destination in ['TJM','KBV'] or segments[0].origin in ['ROV','HKT'] and segments[0].destination in ['ROV','HKT'] or segments[0].origin in ['YKS','DXB'] and segments[0].destination in ['YKS','DXB'] or segments[0].origin in ['MOW','GDX'] and segments[0].destination in ['MOW','GDX'] or segments[0].origin in ['VOG','HKT'] and segments[0].destination in ['VOG','HKT'] or segments[0].origin in ['EGO','BKK'] and segments[0].destination in ['EGO','BKK'] or segments[0].origin in ['KGD','MOW'] and segments[0].destination in ['KGD','MOW'] or segments[0].origin in ['IKT','USM'] and segments[0].destination in ['IKT','USM'] or segments[0].origin in ['MOW','TDX'] and segments[0].destination in ['MOW','TDX'] or segments[0].origin in ['SVX','SGN'] and segments[0].destination in ['SVX','SGN'] or segments[0].origin in ['KRR','VRN'] and segments[0].destination in ['KRR','VRN'] or segments[0].origin in ['BAX','PUJ'] and segments[0].destination in ['BAX','PUJ'] or segments[0].origin in ['HRG','LED'] and segments[0].destination in ['HRG','LED'] or segments[0].origin in ['ASF','BKK'] and segments[0].destination in ['ASF','BKK'] or segments[0].origin in ['ROV','SHJ'] and segments[0].destination in ['ROV','SHJ'] or segments[0].origin in ['KRR','HKT'] and segments[0].destination in ['KRR','HKT'] or segments[0].origin in ['MOW','AYT'] and segments[0].destination in ['MOW','AYT'] or segments[0].origin in ['SVX','HKT'] and segments[0].destination in ['SVX','HKT'] or segments[0].origin in ['SVX','BKK'] and segments[0].destination in ['SVX','BKK'] or segments[0].origin in ['MOW','SGN'] and segments[0].destination in ['MOW','SGN'] or segments[0].origin in ['PEE','SSH'] and segments[0].destination in ['PEE','SSH'] or segments[0].origin in ['MOW','DXB'] and segments[0].destination in ['MOW','DXB'] or segments[0].origin in ['YKS','PUJ'] and segments[0].destination in ['YKS','PUJ'] or segments[0].origin in ['MOW','AUH'] and segments[0].destination in ['MOW','AUH'] or segments[0].origin in ['IKT','UTP'] and segments[0].destination in ['IKT','UTP'] or segments[0].origin in ['ROV','BKK'] and segments[0].destination in ['ROV','BKK'] or segments[0].origin in ['KGD','SGN'] and segments[0].destination in ['KGD','SGN'] or segments[0].origin in ['SCW','HKT'] and segments[0].destination in ['SCW','HKT'] or segments[0].origin in ['KGD','BKK'] and segments[0].destination in ['KGD','BKK'] or segments[0].origin in ['SGC','HKT'] and segments[0].destination in ['SGC','HKT'] or segments[0].origin in ['TSE','SKG'] and segments[0].destination in ['TSE','SKG'] or segments[0].origin in ['PKC','BKK'] and segments[0].destination in ['PKC','BKK'] or segments[0].origin in ['KEJ','KBV'] and segments[0].destination in ['KEJ','KBV'] or segments[0].origin in ['SCW','BKK'] and segments[0].destination in ['SCW','BKK'] or segments[0].origin in ['MOW','AMM'] and segments[0].destination in ['MOW','AMM'] or segments[0].origin in ['KZN','SGN'] and segments[0].destination in ['KZN','SGN'] or segments[0].origin in ['AER','VOG'] and segments[0].destination in ['AER','VOG'] or segments[0].origin in ['SVX','DXB'] and segments[0].destination in ['SVX','DXB'] or segments[0].origin in ['BQS','HKT'] and segments[0].destination in ['BQS','HKT'] or segments[0].origin in ['LED','DXB'] and segments[0].destination in ['LED','DXB'] or segments[0].origin in ['ROV','KBV'] and segments[0].destination in ['ROV','KBV'] or segments[0].origin in ['REN','BKK'] and segments[0].destination in ['REN','BKK'] or segments[0].origin in ['PEE','HKT'] and segments[0].destination in ['PEE','HKT'] or segments[0].origin in ['AER','BKK'] and segments[0].destination in ['AER','BKK'] or segments[0].origin in ['SCW','SGN'] and segments[0].destination in ['SCW','SGN'] or segments[0].origin in ['KUF','SIP'] and segments[0].destination in ['KUF','SIP'] or segments[0].origin in ['KRR','AQJ'] and segments[0].destination in ['KRR','AQJ'] or segments[0].origin in ['MOW','KBV'] and segments[0].destination in ['MOW','KBV'] or segments[0].origin in ['PEE','BKK'] and segments[0].destination in ['PEE','BKK'] or segments[0].origin in ['NBC','DXB'] and segments[0].destination in ['NBC','DXB'] or segments[0].origin in ['UUD','HKT'] and segments[0].destination in ['UUD','HKT'] or segments[0].origin in ['OVB','BTK'] and segments[0].destination in ['OVB','BTK'] or segments[0].origin in ['OVB','USM'] and segments[0].destination in ['OVB','USM'] or segments[0].origin in ['VRA','RTW'] and segments[0].destination in ['VRA','RTW'] or segments[0].origin in ['VOZ','SHJ'] and segments[0].destination in ['VOZ','SHJ'] or segments[0].origin in ['NOZ','DXB'] and segments[0].destination in ['NOZ','DXB'] or segments[0].origin in ['SGC','SGN'] and segments[0].destination in ['SGC','SGN'] or segments[0].origin in ['SVX','AYT'] and segments[0].destination in ['SVX','AYT'] or segments[0].origin in ['KZN','KBV'] and segments[0].destination in ['KZN','KBV'] or segments[0].origin in ['TJM','USM'] and segments[0].destination in ['TJM','USM'] or segments[0].origin in ['ROV','SGN'] and segments[0].destination in ['ROV','SGN'] or segments[0].origin in ['LED','AMM'] and segments[0].destination in ['LED','AMM'] or segments[0].origin in ['PUJ','RTW'] and segments[0].destination in ['PUJ','RTW'] or segments[0].origin in ['NJC','BKK'] and segments[0].destination in ['NJC','BKK'] or segments[0].origin in ['MOW','ROM'] and segments[0].destination in ['MOW','ROM'] or segments[0].origin in ['VOZ','BKK'] and segments[0].destination in ['VOZ','BKK'] or segments[0].origin in ['LED','KRR'] and segments[0].destination in ['LED','KRR'] or segments[0].origin in ['KUF','SGN'] and segments[0].destination in ['KUF','SGN'] or segments[0].origin in ['MRV','SGN'] and segments[0].destination in ['MRV','SGN'] or segments[0].origin in ['SSH','KGD'] and segments[0].destination in ['SSH','KGD'] or segments[0].origin in ['KRR','BKK'] and segments[0].destination in ['KRR','BKK'] or segments[0].origin in ['OVB','DXB'] and segments[0].destination in ['OVB','DXB'] or segments[0].origin in ['MRV','BKK'] and segments[0].destination in ['MRV','BKK'] or segments[0].origin in ['LED','BKK'] and segments[0].destination in ['LED','BKK'] or segments[0].origin in ['LED','SGN'] and segments[0].destination in ['LED','SGN'] or segments[0].origin in ['REN','HKT'] and segments[0].destination in ['REN','HKT'] or segments[0].origin in ['UUD','BKK'] and segments[0].destination in ['UUD','BKK'] or segments[0].origin in ['GOJ','PUJ'] and segments[0].destination in ['GOJ','PUJ'] or segments[0].origin in ['BAX','TDX'] and segments[0].destination in ['BAX','TDX'] or segments[0].origin in ['KJA','UTP'] and segments[0].destination in ['KJA','UTP'] or segments[0].origin in ['KHV','BKK'] and segments[0].destination in ['KHV','BKK'] or segments[0].origin in ['MOW','BER'] and segments[0].destination in ['MOW','BER'] or segments[0].origin in ['KHV','SGN'] and segments[0].destination in ['KHV','SGN'] or segments[0].origin in ['NJC','DXB'] and segments[0].destination in ['NJC','DXB'] or segments[0].origin in ['BAX','USM'] and segments[0].destination in ['BAX','USM'] or segments[0].origin in ['VRN','ROV'] and segments[0].destination in ['VRN','ROV'] or segments[0].origin in ['IKT','BKK'] and segments[0].destination in ['IKT','BKK'] or segments[0].origin in ['SGC','TDX'] and segments[0].destination in ['SGC','TDX'] or segments[0].origin in ['TOF','BKK'] and segments[0].destination in ['TOF','BKK'] or segments[0].origin in ['SVX','SIP'] and segments[0].destination in ['SVX','SIP'] or segments[0].origin in ['HRG','PEE'] and segments[0].destination in ['HRG','PEE'] or segments[0].origin in ['MOW','GVA'] and segments[0].destination in ['MOW','GVA'] or segments[0].origin in ['MOW','MUC'] and segments[0].destination in ['MOW','MUC'] or segments[0].origin in ['BAX','HKT'] and segments[0].destination in ['BAX','HKT'] or segments[0].origin in ['HTA','SGN'] and segments[0].destination in ['HTA','SGN'] or segments[0].origin in ['MOW','SHJ'] and segments[0].destination in ['MOW','SHJ'] or segments[0].origin in ['UFA','KBV'] and segments[0].destination in ['UFA','KBV'] or segments[0].origin in ['MOW','DJE'] and segments[0].destination in ['MOW','DJE'] or segments[0].origin in ['PEE','SGN'] and segments[0].destination in ['PEE','SGN'] or segments[0].origin in ['LED','CUN'] and segments[0].destination in ['LED','CUN'] or segments[0].origin in ['SVX','MOW'] and segments[0].destination in ['SVX','MOW'] or segments[0].origin in ['SGC','KBV'] and segments[0].destination in ['SGC','KBV'] or segments[0].origin in ['GOJ','DXB'] and segments[0].destination in ['GOJ','DXB'] or segments[0].origin in ['BAX','SGN'] and segments[0].destination in ['BAX','SGN'] or segments[0].origin in ['HRG','ROV'] and segments[0].destination in ['HRG','ROV'] or segments[0].origin in ['IKT','TDX'] and segments[0].destination in ['IKT','TDX'] or segments[0].origin in ['UFA','SGN'] and segments[0].destination in ['UFA','SGN'] or segments[0].origin in ['KHV','HKT'] and segments[0].destination in ['KHV','HKT'] or segments[0].origin in ['KRR','SIP'] and segments[0].destination in ['KRR','SIP'] or segments[0].origin in ['VRA','AAQ'] and segments[0].destination in ['VRA','AAQ'] or segments[0].origin in ['BQS','KBV'] and segments[0].destination in ['BQS','KBV'] or segments[0].origin in ['NBC','BKK'] and segments[0].destination in ['NBC','BKK'] or segments[0].origin in ['NOZ','HKT'] and segments[0].destination in ['NOZ','HKT'] or segments[0].origin in ['LED','SIP'] and segments[0].destination in ['LED','SIP'] or segments[0].origin in ['BGY','MOW'] and segments[0].destination in ['BGY','MOW'] or segments[0].origin in ['EGO','HKT'] and segments[0].destination in ['EGO','HKT'] or segments[0].origin in ['OVB','NHA'] and segments[0].destination in ['OVB','NHA'] or segments[0].origin in ['TJM','BKK'] and segments[0].destination in ['TJM','BKK'] or segments[0].origin in ['BAX','DXB'] and segments[0].destination in ['BAX','DXB'] or segments[0].origin in ['CEK','HKT'] and segments[0].destination in ['CEK','HKT'] or segments[0].origin in ['KGD','HKT'] and segments[0].destination in ['KGD','HKT'] or segments[0].origin in ['SVX','USM'] and segments[0].destination in ['SVX','USM'] or segments[0].origin in ['BAX','KBV'] and segments[0].destination in ['BAX','KBV'] or segments[0].origin in ['UFA','SIP'] and segments[0].destination in ['UFA','SIP'] or segments[0].origin in ['LED','NHA'] and segments[0].destination in ['LED','NHA'] or segments[0].origin in ['VOZ','HKT'] and segments[0].destination in ['VOZ','HKT'] or segments[0].origin in ['GOJ','BKK'] and segments[0].destination in ['GOJ','BKK'] or segments[0].origin in ['ASF','HKT'] and segments[0].destination in ['ASF','HKT'] or segments[0].origin in ['AER','SIP'] and segments[0].destination in ['AER','SIP'] or segments[0].origin in ['TJM','SGN'] and segments[0].destination in ['TJM','SGN'] or segments[0].origin in ['MOW','MIL'] and segments[0].destination in ['MOW','MIL'] or segments[0].origin in ['MOW','USM'] and segments[0].destination in ['MOW','USM'] or segments[0].origin in ['IKT','KBV'] and segments[0].destination in ['IKT','KBV'] or segments[0].origin in ['UFA','HKT'] and segments[0].destination in ['UFA','HKT'] or segments[0].origin in ['EGO','UTP'] and segments[0].destination in ['EGO','UTP'] or segments[0].origin in ['HTA','DXB'] and segments[0].destination in ['HTA','DXB'] or segments[0].origin in ['OMS','SGN'] and segments[0].destination in ['OMS','SGN'] or segments[0].origin in ['PUJ','HTA'] and segments[0].destination in ['PUJ','HTA'] or segments[0].origin in ['OMS','HKT'] and segments[0].destination in ['OMS','HKT'] or segments[0].origin in ['KUF','DXB'] and segments[0].destination in ['KUF','DXB'] or segments[0].origin in ['MCX','BKK'] and segments[0].destination in ['MCX','BKK'] or segments[0].origin in ['IKT','HKT'] and segments[0].destination in ['IKT','HKT'] or segments[0].origin in ['ROV','DXB'] and segments[0].destination in ['ROV','DXB'] or segments[0].origin in ['TJM','DXB'] and segments[0].destination in ['TJM','DXB'] or segments[0].origin in ['OMS','USM'] and segments[0].destination in ['OMS','USM'] or segments[0].origin in ['KRR','DXB'] and segments[0].destination in ['KRR','DXB'] or segments[0].origin in ['KZN','HKT'] and segments[0].destination in ['KZN','HKT'] or segments[0].origin in ['LED','USM'] and segments[0].destination in ['LED','USM'] or segments[0].origin in ['MRV','HKT'] and segments[0].destination in ['MRV','HKT'] or segments[0].origin in ['TOF','HKT'] and segments[0].destination in ['TOF','HKT'] or segments[0].origin in ['KEJ','DXB'] and segments[0].destination in ['KEJ','DXB'] or segments[0].origin in ['MRV','DXB'] and segments[0].destination in ['MRV','DXB'] or segments[0].origin in ['OGZ','BKK'] and segments[0].destination in ['OGZ','BKK'] or segments[0].origin in ['PEE','UTP'] and segments[0].destination in ['PEE','UTP'] or segments[0].origin in ['ROV','USM'] and segments[0].destination in ['ROV','USM'] or segments[0].origin in ['KUF','UTP'] and segments[0].destination in ['KUF','UTP'] or segments[0].origin in ['KZN','SIP'] and segments[0].destination in ['KZN','SIP'] or segments[0].origin in ['REN','SHJ'] and segments[0].destination in ['REN','SHJ'] or segments[0].origin in ['KRR','SGN'] and segments[0].destination in ['KRR','SGN'] or segments[0].origin in ['MOW','BUH'] and segments[0].destination in ['MOW','BUH'] or segments[0].origin in ['BQS','SGN'] and segments[0].destination in ['BQS','SGN'] or segments[0].origin in ['TOF','DXB'] and segments[0].destination in ['TOF','DXB'] or segments[0].origin in ['KUF','HKT'] and segments[0].destination in ['KUF','HKT'] or segments[0].origin in ['HRG','SCW'] and segments[0].destination in ['HRG','SCW'] or segments[0].origin in ['PEE','AYT'] and segments[0].destination in ['PEE','AYT'] or segments[0].origin in ['OMS','BKK'] and segments[0].destination in ['OMS','BKK'] or segments[0].origin in ['CEK','BKK'] and segments[0].destination in ['CEK','BKK'] or segments[0].origin in ['YKS','HKT'] and segments[0].destination in ['YKS','HKT'] or segments[0].origin in ['MOW','TRN'] and segments[0].destination in ['MOW','TRN'] or segments[0].origin in ['OMS','DXB'] and segments[0].destination in ['OMS','DXB'] or segments[0].origin in ['VVO','HKT'] and segments[0].destination in ['VVO','HKT'] or segments[0].origin in ['KZN','USM'] and segments[0].destination in ['KZN','USM'] or segments[0].origin in ['MOW','POP'] and segments[0].destination in ['MOW','POP'] or segments[0].origin in ['LED','UTP'] and segments[0].destination in ['LED','UTP'] or segments[0].origin in ['GOJ','SGN'] and segments[0].destination in ['GOJ','SGN'] or segments[0].origin in ['VVO','USM'] and segments[0].destination in ['VVO','USM'] or segments[0].origin in ['PUJ','CEE'] and segments[0].destination in ['PUJ','CEE'] or segments[0].origin in ['SVX','PRG'] and segments[0].destination in ['SVX','PRG'] or segments[0].origin in ['VRA','ABA'] and segments[0].destination in ['VRA','ABA'] or segments[0].origin in ['IKT','SGN'] and segments[0].destination in ['IKT','SGN'] or segments[0].origin in ['VOG','SHJ'] and segments[0].destination in ['VOG','SHJ'])`
	program, err := jsexpr.Compile(expression, jsexpr.TypeCheck(env))
	if err != nil {
		b.Fatal(err)
	}

	var out interface{}
	virtualMachine := vm.VM{}
	virtualMachine.Init(program, env)
	for n := 0; n < b.N; n++ {
		out, err = virtualMachine.Run(program, env)
	}
	if err != nil {
		b.Fatal(err)
	}
	if !out.(bool) {
		b.Fail()
	}
}

type Env struct {
	Tweets []Tweet
}

type Tweet struct {
	Text string
	Date time.Time
}

func (Env) Format(t time.Time) string { return t.Format(time.RFC822) }

func BenchmarkCompileWithoutEnv(b *testing.B) {
	code := `map(filter(tweets, {len(.text) > 0}), {.text + format(.date)})`
	env := Env{
		Tweets: []Tweet{{"Oh My God!", time.Now()}, {"How you doin?", time.Now()}, {"Could I be wearing any more clothes?", time.Now()}},
	}

	program, _ := jsexpr.Compile(code)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, _ = jsexpr.Run(program, env)
	}
	b.StopTimer()
}

func BenchmarkCompileWithEnv(b *testing.B) {
	code := `map(filter(tweets, {len(.text) > 0}), {.text + format(.date)})`
	env := Env{
		Tweets: []Tweet{{"Oh My God!", time.Now()}, {"How you doin?", time.Now()}, {"Could I be wearing any more clothes?", time.Now()}},
	}

	program, _ := jsexpr.Compile(code, jsexpr.TypeCheck(Env{}))

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, _ = jsexpr.Run(program, env)
	}
	b.StopTimer()
}
