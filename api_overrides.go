/*
The MIT License (MIT)

Copyright (c) 2016 isaac dawson

Permission is hereby granted, free of charge, to any person obtaining a copy
of this software and associated documentation files (the "Software"), to deal
in the Software without restriction, including without limitation the rights
to use, copy, modify, merge, publish, distribute, sublicense, and/or sell
copies of the Software, and to permit persons to whom the Software is
furnished to do so, subject to the following conditions:

The above copyright notice and this permission notice shall be included in
all copies or substantial portions of the Software.

THE SOFTWARE IS PROVIDED "AS IS", WITHOUT WARRANTY OF ANY KIND, EXPRESS OR
IMPLIED, INCLUDING BUT NOT LIMITED TO THE WARRANTIES OF MERCHANTABILITY,
FITNESS FOR A PARTICULAR PURPOSE AND NONINFRINGEMENT. IN NO EVENT SHALL THE
AUTHORS OR COPYRIGHT HOLDERS BE LIABLE FOR ANY CLAIM, DAMAGES OR OTHER
LIABILITY, WHETHER IN AN ACTION OF CONTRACT, TORT OR OTHERWISE, ARISING FROM,
OUT OF OR IN CONNECTION WITH THE SOFTWARE OR THE USE OR OTHER DEALINGS IN
THE SOFTWARE.
*/

package autogcd

import (
	"encoding/json"
	"github.com/wirepair/gcd"
	"github.com/wirepair/gcd/gcdapi"
	"github.com/wirepair/gcd/gcdmessage"
)

/*
This is for overriding gcdapi calls that for some reason or another need parameters removed or they won't
work.
*/

// Evaluate - Evaluates expression on global object.
// This method is overriden because the docs lie, passing 0 returns invalid context id, we must remove it from the map
// entirely for the call to work on the global object.
// expression - Expression to evaluate.
// objectGroup - Symbolic group name that can be used to release multiple objects.
// includeCommandLineAPI - Determines whether Command Line API should be available during the evaluation.
// doNotPauseOnExceptionsAndMuteConsole - Specifies whether evaluation should stop on exceptions and mute console. Overrides setPauseOnException state.
// contextId - Specifies in which isolated context to perform evaluation. Each content script lives in an isolated context and this parameter may be used to specify one of those contexts. If the parameter is omitted or 0 the evaluation will be performed in the context of the inspected page.
// returnByValue - Whether the result is expected to be a JSON object that should be sent by value.
// generatePreview - Whether preview should be generated for the result.
// Returns -  result - Evaluation result. wasThrown - True if the result was thrown during the evaluation. exceptionDetails - Exception details.
func overridenRuntimeEvaluate(target *gcd.ChromeTarget, expression string, objectGroup string, includeCommandLineAPI bool, doNotPauseOnExceptionsAndMuteConsole bool, contextId int, returnByValue bool, generatePreview bool) (*gcdapi.RuntimeRemoteObject, bool, *gcdapi.RuntimeExceptionDetails, error) {
	paramRequest := make(map[string]interface{}, 7)
	paramRequest["expression"] = expression
	paramRequest["objectGroup"] = objectGroup
	paramRequest["includeCommandLineAPI"] = includeCommandLineAPI
	paramRequest["doNotPauseOnExceptionsAndMuteConsole"] = doNotPauseOnExceptionsAndMuteConsole
	// only add context id if it's non-zero
	if contextId != 0 {
		paramRequest["contextId"] = contextId
	}
	paramRequest["returnByValue"] = returnByValue
	paramRequest["generatePreview"] = generatePreview
	resp, err := gcdmessage.SendCustomReturn(target, target.GetSendCh(), &gcdmessage.ParamRequest{Id: target.GetId(), Method: "Runtime.evaluate", Params: paramRequest})
	if err != nil {
		return nil, false, nil, err
	}

	if resp == nil {
		return nil, false, nil, &gcdmessage.ChromeEmptyResponseErr{}
	}

	var chromeData struct {
		Result struct {
			Result           *gcdapi.RuntimeRemoteObject
			WasThrown        bool
			ExceptionDetails *gcdapi.RuntimeExceptionDetails
		}
	}
	// test if error first
	cerr := &gcdmessage.ChromeErrorResponse{}
	json.Unmarshal(resp.Data, cerr)
	if cerr != nil && cerr.Error != nil {
		return nil, false, nil, &gcdmessage.ChromeRequestErr{Resp: cerr}
	}

	if err := json.Unmarshal(resp.Data, &chromeData); err != nil {
		return nil, false, nil, err
	}

	return chromeData.Result.Result, chromeData.Result.WasThrown, chromeData.Result.ExceptionDetails, nil
}
