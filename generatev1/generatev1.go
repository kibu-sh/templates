package main

import (
	"bytes"
	"fmt"
	"os"
	"os/exec"

	"github.com/dave/jennifer/jen"
)

// Main function to generate the code.
func main() {
	code := GenerateCode()

	// Render the code to a buffer
	var buf bytes.Buffer
	err := code.Render(&buf)
	if err != nil {
		fmt.Printf("Error rendering code: %v\n", err)
		os.Exit(1)
	}

	// Define the output filename
	filename := "generated.go"

	// Write the buffer to the file
	err = os.WriteFile(filename, buf.Bytes(), 0644)
	if err != nil {
		fmt.Printf("Error writing file: %v\n", err)
		os.Exit(1)
	}

	// Run 'go fmt' on the generated file
	cmd := exec.Command("go", "fmt", filename)
	err = cmd.Run()
	if err != nil {
		fmt.Printf("Error running go fmt: %v\n", err)
		os.Exit(1)
	}
}

// GenerateCode generates the entire Go file using jennifer.
func GenerateCode() *jen.File {
	f := jen.NewFile("billingv1")

	// Add imports.
	AddImports(f)

	// Add constants.
	f.Add(GenerateConstants())

	// Add type aliases.
	f.Add(GenerateTypeAliases())

	// Add GetHandleOpts struct.
	f.Add(GenerateGetHandleOpts())

	// Add interfaces and structs implementing them.
	f.Add(GenerateInterfacesAndStructs())

	// Implementing WorkflowsClient.
	f.Add(GenerateWorkflowsClient())

	// Implementing WorkflowsProxy.
	f.Add(GenerateWorkflowsProxy())

	// Implementing CustomerBillingWorkflowClient.
	f.Add(GenerateCustomerBillingWorkflowClient())

	// Implementing CustomerBillingWorkflowChildClient.
	f.Add(GenerateCustomerBillingWorkflowChildClient())

	// Implementing CustomerBillingExternalRun.
	f.Add(GenerateCustomerBillingExternalRun())

	// Implementing CustomerBillingWorkflowRun.
	f.Add(GenerateCustomerBillingWorkflowRun())

	// Implementing CustomerBillingWorkflowChildRun.
	f.Add(GenerateCustomerBillingWorkflowChildRun())

	return f
}

// AddImports adds necessary imports to the file.
func AddImports(f *jen.File) {
	f.ImportName("context", "context")
	f.ImportName("github.com/kibu-sh/kibu/pkg/transport/temporal", "temporal")
	f.ImportName("go.temporal.io/sdk/client", "client")
	f.ImportName("go.temporal.io/sdk/workflow", "workflow")
}

// GenerateConstants generates the constant declarations.
func GenerateConstants() *jen.Statement {
	return jen.Const().Defs(
		jen.Id("barv1ServiceWatchBillingBillingName").Op("=").Lit("barv1.WatchBillingBilling"),
		jen.Id("barv1CustomerBillingWorkflowName").Op("=").Lit("barv1.customerBillingWorkflow"),
		jen.Id("barv1CustomerBillingWorkflowAttemptPaymentName").Op("=").Lit("barv1.customerBillingWorkflow.AttemptPayment"),
		jen.Id("barv1CustomerBillingWorkflowGetAccountDetailsName").Op("=").Lit("barv1.customerBillingWorkflow.GetAccountDetails"),
		jen.Id("barv1CustomerBillingWorkflowCancelBillingName").Op("=").Lit("barv1.customerBillingWorkflow.CancelBilling"),
		jen.Id("barv1CustomerBillingWorkflowSetDiscountName").Op("=").Lit("barv1.customerBillingWorkflow.SetDiscount"),
		jen.Id("barv1ActivitiesChargePaymentMethodName").Op("=").Lit("barv1.activities.ChargePaymentMethod"),
	)
}

// GenerateTypeAliases generates type aliases.
func GenerateTypeAliases() *jen.Statement {
	return jen.Empty().
		Type().Id("WorkflowOptionFunc").Op("=").Qual("github.com/kibu-sh/kibu/pkg/transport/temporal", "WorkflowOptionFunc").
		Type().Id("ActivityOptionFunc").Op("=").Qual("github.com/kibu-sh/kibu/pkg/transport/temporal", "ActivityOptionFunc").
		Type().Id("UpdateOptionFunc").Op("=").Qual("github.com/kibu-sh/kibu/pkg/transport/temporal", "UpdateOptionFunc")
}

// GenerateGetHandleOpts generates the GetHandleOpts struct.
func GenerateGetHandleOpts() *jen.Statement {
	return jen.Type().Id("GetHandleOpts").Struct(
		jen.Id("WorkflowID").String(),
		jen.Id("RunID").String(),
	)
}

// GenerateInterfacesAndStructs generates all interfaces and their implementations.
func GenerateInterfacesAndStructs() *jen.Statement {
	return jen.Empty().
		Add(GenerateCustomerBillingWorkflowRunInterface()).
		Add(GenerateCustomerBillingWorkflowClientInterface()).
		Add(GenerateCustomerBillingWorkflowChildRunInterface()).
		Add(GenerateCustomerBillingExternalRunInterface()).
		Add(GenerateCustomerBillingWorkflowChildClientInterface()).
		Add(GenerateWorkflowsProxyInterface()).
		Add(GenerateWorkflowsClientInterface())
}

// GenerateCustomerBillingWorkflowRunInterface generates the CustomerBillingWorkflowRun interface.
func GenerateCustomerBillingWorkflowRunInterface() *jen.Statement {
	return jen.Type().Id("CustomerBillingWorkflowRun").Interface(
		jen.Id("ID").Params().String(),
		jen.Id("RunID").Params().String(),
		jen.Line(),
		jen.Id("Get").Params(jen.Id("ctx").Qual("context", "Context")).Params(
			jen.Id("CustomerBillingResponse"),
			jen.Error(),
		),
		jen.Line(),
		jen.Id("AttemptPayment").Params(
			jen.Id("ctx").Qual("context", "Context"),
			jen.Id("req").Id("AttemptPaymentRequest"),
			jen.Id("mods").Op("...").Id("WorkflowOptionFunc"),
		).Params(
			jen.Id("AttemptPaymentResponse"),
			jen.Error(),
		),
		jen.Line(),
		jen.Id("GetAccountDetails").Params(
			jen.Id("ctx").Qual("context", "Context"),
			jen.Id("req").Id("GetAccountDetailsRequest"),
			jen.Id("mods").Op("...").Id("WorkflowOptionFunc"),
		).Params(
			jen.Id("GetAccountDetailsResult"),
			jen.Error(),
		),
		jen.Line(),
		jen.Id("CancelBilling").Params(
			jen.Id("ctx").Qual("context", "Context"),
			jen.Id("req").Id("CancelBillingSignal"),
			jen.Id("mods").Op("...").Id("WorkflowOptionFunc"),
		).Error(),
		jen.Id("SetDiscount").Params(
			jen.Id("ctx").Qual("context", "Context"),
			jen.Id("req").Id("SetDiscountSignal"),
			jen.Id("mods").Op("...").Id("WorkflowOptionFunc"),
		).Error(),
	)
}

// GenerateCustomerBillingWorkflowClientInterface generates the CustomerBillingWorkflowClient interface.
func GenerateCustomerBillingWorkflowClientInterface() *jen.Statement {
	return jen.Type().Id("CustomerBillingWorkflowClient").Interface(
		jen.Id("Execute").Params(
			jen.Id("ctx").Qual("context", "Context"),
			jen.Id("req").Id("CustomerBillingRequest"),
		).Params(
			jen.Id("CustomerBillingWorkflowRun"),
			jen.Error(),
		),
		jen.Id("GetHandle").Params(
			jen.Id("ctx").Qual("context", "Context"),
			jen.Id("ref").Id("GetHandleOpts"),
		).Params(
			jen.Id("CustomerBillingWorkflowRun"),
			jen.Error(),
		),
		jen.Id("ExecuteWithSetDiscount").Params(
			jen.Id("ctx").Qual("context", "Context"),
			jen.Id("req").Id("SetDiscountSignal"),
		).Params(
			jen.Id("CustomerBillingWorkflowRun"),
			jen.Error(),
		),
	)
}

// GenerateCustomerBillingWorkflowChildRunInterface generates the CustomerBillingWorkflowChildRun interface.
func GenerateCustomerBillingWorkflowChildRunInterface() *jen.Statement {
	return jen.Type().Id("CustomerBillingWorkflowChildRun").Interface(
		jen.Id("ID").Params().String(),
		jen.Id("IsReady").Params().Bool(),
		jen.Id("Underlying").Params().Qual("go.temporal.io/sdk/workflow", "ChildWorkflowFuture"),
		jen.Id("Get").Params(
			jen.Id("ctx").Qual("go.temporal.io/sdk/workflow", "Context"),
		).Params(
			jen.Id("CustomerBillingResponse"),
			jen.Error(),
		),
		jen.Id("CancelBilling").Params(
			jen.Id("ctx").Qual("go.temporal.io/sdk/workflow", "Context"),
			jen.Id("req").Id("CancelBillingSignal"),
		).Error(),
		jen.Id("SetDiscount").Params(
			jen.Id("ctx").Qual("go.temporal.io/sdk/workflow", "Context"),
			jen.Id("req").Id("SetDiscountSignal"),
		).Error(),
	)
}

// GenerateCustomerBillingExternalRunInterface generates the CustomerBillingExternalRun interface.
func GenerateCustomerBillingExternalRunInterface() *jen.Statement {
	return jen.Type().Id("CustomerBillingExternalRun").Interface(
		jen.Id("ID").Params().String(),
		jen.Id("RunID").Params().String(),
		jen.Line(),
		jen.Id("RequestCancellation").Params(
			jen.Id("ctx").Qual("go.temporal.io/sdk/workflow", "Context"),
		).Error(),
		jen.Line(),
		jen.Id("CancelBilling").Params(
			jen.Id("ctx").Qual("go.temporal.io/sdk/workflow", "Context"),
			jen.Id("req").Id("CancelBillingSignal"),
		).Error(),
		jen.Id("CancelBillingAsync").Params(
			jen.Id("ctx").Qual("go.temporal.io/sdk/workflow", "Context"),
			jen.Id("req").Id("CancelBillingSignal"),
		).Params(
			jen.Qual("go.temporal.io/sdk/workflow", "Future"),
		),
		jen.Line(),
		jen.Id("SetDiscount").Params(
			jen.Id("ctx").Qual("go.temporal.io/sdk/workflow", "Context"),
			jen.Id("req").Id("SetDiscountSignal"),
		).Error(),
		jen.Id("SetDiscountAsync").Params(
			jen.Id("ctx").Qual("go.temporal.io/sdk/workflow", "Context"),
			jen.Id("req").Id("SetDiscountSignal"),
		).Params(
			jen.Qual("go.temporal.io/sdk/workflow", "Future"),
		),
	)
}

// GenerateCustomerBillingWorkflowChildClientInterface generates the CustomerBillingWorkflowChildClient interface.
func GenerateCustomerBillingWorkflowChildClientInterface() *jen.Statement {
	return jen.Type().Id("CustomerBillingWorkflowChildClient").Interface(
		jen.Id("Execute").Params(
			jen.Id("ctx").Qual("go.temporal.io/sdk/workflow", "Context"),
			jen.Id("req").Id("CustomerBillingRequest"),
		).Params(
			jen.Id("CustomerBillingResponse"),
			jen.Error(),
		),
		jen.Id("ExecuteAsync").Params(
			jen.Id("ctx").Qual("go.temporal.io/sdk/workflow", "Context"),
			jen.Id("req").Id("CustomerBillingRequest"),
		).Params(
			jen.Id("CustomerBillingWorkflowChildRun"),
		),
		jen.Id("External").Params(
			jen.Id("ref").Id("GetHandleOpts"),
		).Params(
			jen.Id("CustomerBillingExternalRun"),
		),
	)
}

// GenerateWorkflowsProxyInterface generates the WorkflowsProxy interface.
func GenerateWorkflowsProxyInterface() *jen.Statement {
	return jen.Type().Id("WorkflowsProxy").Interface(
		jen.Id("CustomerSubscriptions").Params().Id("CustomerBillingWorkflowChildClient"),
	)
}

// GenerateWorkflowsClientInterface generates the WorkflowsClient interface.
func GenerateWorkflowsClientInterface() *jen.Statement {
	return jen.Type().Id("WorkflowsClient").Interface(
		jen.Id("CustomerSubscriptions").Params().Id("CustomerBillingWorkflowClient"),
	)
}

// GenerateWorkflowsClient generates the workflowsClient struct and its methods.
func GenerateWorkflowsClient() *jen.Statement {
	return jen.Type().Id("workflowsClient").Struct(
		jen.Id("client").Qual("go.temporal.io/sdk/client", "Client"),
	).Line().
		Add(jen.Func().Params(jen.Id("w").Op("*").Id("workflowsClient")).Id("CustomerSubscriptions").Params().Id("CustomerBillingWorkflowClient").Block(
			jen.Return(jen.Op("&").Id("customerBillingWorkflowClient").Values(
				jen.Dict{
					jen.Id("client"): jen.Id("w").Dot("client"),
				},
			)),
		))
}

// GenerateWorkflowsProxy generates the workflowsProxy struct and its methods.
func GenerateWorkflowsProxy() *jen.Statement {
	return jen.Type().Id("workflowsProxy").Struct().Line().
		Add(jen.Func().Params(jen.Id("w").Op("*").Id("workflowsProxy")).Id("CustomerSubscriptions").Params().Id("CustomerBillingWorkflowChildClient").Block(
			jen.Return(jen.Op("&").Id("customerBillingWorkflowChildClient").Values()),
		))
}

// GenerateCustomerBillingWorkflowClient generates the customerBillingWorkflowClient struct and methods.
func GenerateCustomerBillingWorkflowClient() *jen.Statement {
	return jen.Type().Id("customerBillingWorkflowClient").Struct(
		jen.Id("client").Qual("go.temporal.io/sdk/client", "Client"),
	).Line().
		Add(jen.Func().Params(jen.Id("c").Op("*").Id("customerBillingWorkflowClient")).Id("Execute").Params(
			jen.Id("ctx").Qual("context", "Context"),
			jen.Id("req").Id("CustomerBillingRequest"),
		).Params(
			jen.Id("CustomerBillingWorkflowRun"),
			jen.Error(),
		).Block(
			jen.Id("options").Op(":=").Qual("github.com/kibu-sh/kibu/pkg/transport/temporal", "NewWorkflowOptionsBuilder").Call().
				Dot("WithTaskQueue").Call(jen.Lit("default")).
				Dot("WithProvidersWhenSupported").Call().
				Dot("AsStartOptions").Call(),
			jen.Line(),
			jen.List(jen.Id("we"), jen.Err()).Op(":=").Id("c").Dot("client").Dot("ExecuteWorkflow").Call(
				jen.Id("ctx"),
				jen.Id("options"),
				jen.Id("barv1CustomerBillingWorkflowName"),
				jen.Id("req"),
			),
			jen.If(jen.Err().Op("!=").Nil()).Block(
				jen.Return(jen.Nil(), jen.Err()),
			),
			jen.Line(),
			jen.Return(jen.Op("&").Id("customerBillingWorkflowRun").Values(
				jen.Dict{
					jen.Id("client"):      jen.Id("c").Dot("client"),
					jen.Id("workflowRun"): jen.Id("we"),
				},
			), jen.Nil()),
		)).
		Add(jen.Func().Params(jen.Id("c").Op("*").Id("customerBillingWorkflowClient")).Id("GetHandle").Params(
			jen.Id("ctx").Qual("context", "Context"),
			jen.Id("ref").Id("GetHandleOpts"),
		).Params(
			jen.Id("CustomerBillingWorkflowRun"),
			jen.Error(),
		).Block(
			jen.Id("we").Op(":=").Id("c").Dot("client").Dot("GetWorkflow").Call(
				jen.Id("ctx"),
				jen.Id("ref").Dot("WorkflowID"),
				jen.Id("ref").Dot("RunID"),
			),
			jen.Return(jen.Op("&").Id("customerBillingWorkflowRun").Values(
				jen.Dict{
					jen.Id("client"):      jen.Id("c").Dot("client"),
					jen.Id("workflowRun"): jen.Id("we"),
				},
			), jen.Nil()),
		)).
		Add(jen.Func().Params(jen.Id("c").Op("*").Id("customerBillingWorkflowClient")).Id("ExecuteWithSetDiscount").Params(
			jen.Id("ctx").Qual("context", "Context"),
			jen.Id("req").Id("SetDiscountSignal"),
		).Params(
			jen.Id("CustomerBillingWorkflowRun"),
			jen.Error(),
		).Block(
			jen.Id("options").Op(":=").Qual("github.com/kibu-sh/kibu/pkg/transport/temporal", "NewWorkflowOptionsBuilder").Call().
				Dot("WithTaskQueue").Call(jen.Lit("default")).
				Dot("WithProvidersWhenSupported").Call().
				Dot("AsStartOptions").Call(),
			jen.Line(),
			jen.List(jen.Id("we"), jen.Err()).Op(":=").Id("c").Dot("client").Dot("SignalWithStartWorkflow").Call(
				jen.Id("ctx"),
				jen.Id("options").Dot("ID"),
				jen.Id("barv1CustomerBillingWorkflowSetDiscountName"),
				jen.Id("req"),
				jen.Id("options"),
				jen.Id("barv1CustomerBillingWorkflowName"),
				jen.Id("req"),
			),
			jen.If(jen.Err().Op("!=").Nil()).Block(
				jen.Return(jen.Nil(), jen.Err()),
			),
			jen.Line(),
			jen.Return(jen.Op("&").Id("customerBillingWorkflowRun").Values(
				jen.Dict{
					jen.Id("client"):      jen.Id("c").Dot("client"),
					jen.Id("workflowRun"): jen.Id("we"),
				},
			), jen.Nil()),
		))
}

// GenerateCustomerBillingWorkflowChildClient generates the customerBillingWorkflowChildClient struct and methods.
func GenerateCustomerBillingWorkflowChildClient() *jen.Statement {
	return jen.Type().Id("customerBillingWorkflowChildClient").Struct().Line().
		Add(jen.Func().Params(jen.Id("c").Op("*").Id("customerBillingWorkflowChildClient")).Id("Execute").Params(
			jen.Id("ctx").Qual("go.temporal.io/sdk/workflow", "Context"),
			jen.Id("req").Id("CustomerBillingRequest"),
		).Params(
			jen.Id("CustomerBillingResponse"),
			jen.Error(),
		).Block(
			jen.Id("future").Op(":=").Id("c").Dot("ExecuteAsync").Call(
				jen.Id("ctx"),
				jen.Id("req"),
			),
			jen.Return(jen.Id("future").Dot("Get").Call(jen.Id("ctx"))),
		)).
		Add(jen.Func().Params(jen.Id("c").Op("*").Id("customerBillingWorkflowChildClient")).Id("ExecuteAsync").Params(
			jen.Id("ctx").Qual("go.temporal.io/sdk/workflow", "Context"),
			jen.Id("req").Id("CustomerBillingRequest"),
		).Params(
			jen.Id("CustomerBillingWorkflowChildRun"),
		).Block(
			jen.Id("options").Op(":=").Qual("github.com/kibu-sh/kibu/pkg/transport/temporal", "NewWorkflowOptionsBuilder").Call().
				Dot("WithTaskQueue").Call(jen.Lit("default")).
				Dot("WithProvidersWhenSupported").Call().
				Dot("AsChildOptions").Call(),
			jen.Line(),
			jen.Id("ctx").Op("=").Qual("go.temporal.io/sdk/workflow", "WithChildOptions").Call(
				jen.Id("ctx"),
				jen.Id("options"),
			),
			jen.Id("childFuture").Op(":=").Qual("go.temporal.io/sdk/workflow", "ExecuteChildWorkflow").Call(
				jen.Id("ctx"),
				jen.Id("barv1CustomerBillingWorkflowName"),
				jen.Id("req"),
			),
			jen.Line(),
			jen.Return(jen.Op("&").Id("customerBillingWorkflowChildRun").Values(
				jen.Dict{
					jen.Id("childFuture"): jen.Id("childFuture"),
				},
			)),
		)).
		Add(jen.Func().Params(jen.Id("c").Op("*").Id("customerBillingWorkflowChildClient")).Id("External").Params(
			jen.Id("ref").Id("GetHandleOpts"),
		).Params(
			jen.Id("CustomerBillingExternalRun"),
		).Block(
			jen.Return(jen.Op("&").Id("customerBillingExternalRun").Values(
				jen.Dict{
					jen.Id("workflowID"): jen.Id("ref").Dot("WorkflowID"),
					jen.Id("runID"):      jen.Id("ref").Dot("RunID"),
				},
			)),
		))
}

// GenerateCustomerBillingExternalRun generates the customerBillingExternalRun struct and methods.
func GenerateCustomerBillingExternalRun() *jen.Statement {
	return jen.Type().Id("customerBillingExternalRun").Struct(
		jen.Id("workflowID").String(),
		jen.Id("runID").String(),
	).Line().
		Add(jen.Func().Params(jen.Id("c").Op("*").Id("customerBillingExternalRun")).Id("ID").Params().String().Block(
			jen.Return(jen.Id("c").Dot("workflowID")),
		)).
		Add(jen.Func().Params(jen.Id("c").Op("*").Id("customerBillingExternalRun")).Id("RunID").Params().String().Block(
			jen.Return(jen.Id("c").Dot("runID")),
		)).
		Add(jen.Func().Params(jen.Id("c").Op("*").Id("customerBillingExternalRun")).Id("RequestCancellation").Params(
			jen.Id("ctx").Qual("go.temporal.io/sdk/workflow", "Context"),
		).Error().Block(
			jen.Return(
				jen.Qual("go.temporal.io/sdk/workflow", "RequestCancelExternalWorkflow").Call(
					jen.Id("ctx"),
					jen.Id("c").Dot("workflowID"),
					jen.Id("c").Dot("runID"),
				).Dot("Get").Call(jen.Id("ctx"), jen.Nil()),
			),
		)).
		// Implement CancelBilling and SetDiscount methods similarly.
		Add(jen.Func().Params(jen.Id("c").Op("*").Id("customerBillingExternalRun")).Id("CancelBilling").Params(
			jen.Id("ctx").Qual("go.temporal.io/sdk/workflow", "Context"),
			jen.Id("req").Id("CancelBillingSignal"),
		).Error().Block(
			jen.Return(
				jen.Id("c").Dot("CancelBillingAsync").Call(jen.Id("ctx"), jen.Id("req")).Dot("Get").Call(jen.Id("ctx"), jen.Nil()),
			),
		)).
		Add(jen.Func().Params(jen.Id("c").Op("*").Id("customerBillingExternalRun")).Id("CancelBillingAsync").Params(
			jen.Id("ctx").Qual("go.temporal.io/sdk/workflow", "Context"),
			jen.Id("req").Id("CancelBillingSignal"),
		).Params(
			jen.Qual("go.temporal.io/sdk/workflow", "Future"),
		).Block(
			jen.Return(
				jen.Qual("go.temporal.io/sdk/workflow", "SignalExternalWorkflow").Call(
					jen.Id("ctx"),
					jen.Id("c").Dot("workflowID"),
					jen.Id("c").Dot("runID"),
					jen.Id("barv1CustomerBillingWorkflowCancelBillingName"),
					jen.Id("req"),
				),
			),
		)).
		// Implement SetDiscount and SetDiscountAsync similarly.
		Add(jen.Func().Params(jen.Id("c").Op("*").Id("customerBillingExternalRun")).Id("SetDiscount").Params(
			jen.Id("ctx").Qual("go.temporal.io/sdk/workflow", "Context"),
			jen.Id("req").Id("SetDiscountSignal"),
		).Error().Block(
			jen.Return(
				jen.Id("c").Dot("SetDiscountAsync").Call(jen.Id("ctx"), jen.Id("req")).Dot("Get").Call(jen.Id("ctx"), jen.Nil()),
			),
		)).
		Add(jen.Func().Params(jen.Id("c").Op("*").Id("customerBillingExternalRun")).Id("SetDiscountAsync").Params(
			jen.Id("ctx").Qual("go.temporal.io/sdk/workflow", "Context"),
			jen.Id("req").Id("SetDiscountSignal"),
		).Params(
			jen.Qual("go.temporal.io/sdk/workflow", "Future"),
		).Block(
			jen.Return(
				jen.Qual("go.temporal.io/sdk/workflow", "SignalExternalWorkflow").Call(
					jen.Id("ctx"),
					jen.Id("c").Dot("workflowID"),
					jen.Id("c").Dot("runID"),
					jen.Id("barv1CustomerBillingWorkflowSetDiscountName"),
					jen.Id("req"),
				),
			),
		))
}

// GenerateCustomerBillingWorkflowRun generates the customerBillingWorkflowRun struct and methods.
func GenerateCustomerBillingWorkflowRun() *jen.Statement {
	return jen.Type().Id("customerBillingWorkflowRun").Struct(
		jen.Id("client").Qual("go.temporal.io/sdk/client", "Client"),
		jen.Id("workflowRun").Qual("go.temporal.io/sdk/client", "WorkflowRun"),
	).Line().
		Add(jen.Func().Params(jen.Id("c").Op("*").Id("customerBillingWorkflowRun")).Id("ID").Params().String().Block(
			jen.Return(jen.Id("c").Dot("workflowRun").Dot("GetID").Call()),
		)).
		Add(jen.Func().Params(jen.Id("c").Op("*").Id("customerBillingWorkflowRun")).Id("RunID").Params().String().Block(
			jen.Return(jen.Id("c").Dot("workflowRun").Dot("GetRunID").Call()),
		)).
		Add(jen.Func().Params(jen.Id("c").Op("*").Id("customerBillingWorkflowRun")).Id("Get").Params(
			jen.Id("ctx").Qual("context", "Context"),
		).Params(
			jen.Id("CustomerBillingResponse"),
			jen.Error(),
		).Block(
			jen.Var().Id("result").Id("CustomerBillingResponse"),
			jen.Err().Op(":=").Id("c").Dot("workflowRun").Dot("Get").Call(jen.Id("ctx"), jen.Op("&").Id("result")),
			jen.Return(jen.Id("result"), jen.Err()),
		)).
		// Implement AttemptPayment, GetAccountDetails, CancelBilling, SetDiscount methods similarly.
		Add(GenerateCustomerBillingWorkflowRunMethods())
}

// GenerateCustomerBillingWorkflowRunMethods generates methods for customerBillingWorkflowRun.
func GenerateCustomerBillingWorkflowRunMethods() *jen.Statement {
	return jen.Empty().
		Add(jen.Func().Params(jen.Id("c").Op("*").Id("customerBillingWorkflowRun")).Id("AttemptPayment").Params(
			jen.Id("ctx").Qual("context", "Context"),
			jen.Id("req").Id("AttemptPaymentRequest"),
			jen.Id("mods").Op("...").Id("WorkflowOptionFunc"),
		).Params(
			jen.Id("AttemptPaymentResponse"),
			jen.Error(),
		).Block(
			jen.Id("options").Op(":=").Qual("github.com/kibu-sh/kibu/pkg/transport/temporal", "NewUpdateOptionsBuilder").Call().
				Dot("WithProvidersWhenSupported").Call(jen.Id("req")).
				Dot("WithWorkflowID").Call(jen.Id("c").Dot("ID").Call()).
				Dot("WithRunID").Call(jen.Id("c").Dot("RunID").Call()).
				Dot("Build").Call(),
			jen.Line(),
			jen.List(jen.Id("updateHandle"), jen.Err()).Op(":=").Id("c").Dot("client").Dot("UpdateWorkflow").Call(
				jen.Id("ctx"),
				jen.Id("options"),
			),
			jen.If(jen.Err().Op("!=").Nil()).Block(
				jen.Return(jen.Id("AttemptPaymentResponse").Values(), jen.Err()),
			),
			jen.Line(),
			jen.Var().Id("result").Id("AttemptPaymentResponse"),
			jen.Err().Op("=").Id("updateHandle").Dot("Get").Call(jen.Id("ctx"), jen.Op("&").Id("result")),
			jen.Return(jen.Id("result"), jen.Err()),
		)).
		// Implement GetAccountDetails, CancelBilling, SetDiscount methods similarly.
		Add(jen.Func().Params(jen.Id("c").Op("*").Id("customerBillingWorkflowRun")).Id("GetAccountDetails").Params(
			jen.Id("ctx").Qual("context", "Context"),
			jen.Id("req").Id("GetAccountDetailsRequest"),
			jen.Id("mods").Op("...").Id("WorkflowOptionFunc"),
		).Params(
			jen.Id("GetAccountDetailsResult"),
			jen.Error(),
		).Block(
			jen.Id("queryResponse"), jen.Err().Op(":=").Id("c").Dot("client").Dot("QueryWorkflow").Call(
				jen.Id("ctx"),
				jen.Id("c").Dot("ID").Call(),
				jen.Id("c").Dot("RunID").Call(),
				jen.Id("barv1CustomerBillingWorkflowGetAccountDetailsName"),
				jen.Id("req"),
			),
			jen.If(jen.Err().Op("!=").Nil()).Block(
				jen.Return(jen.Id("GetAccountDetailsResult").Values(), jen.Err()),
			),
			jen.Line(),
			jen.Var().Id("result").Id("GetAccountDetailsResult"),
			jen.Err().Op("=").Id("queryResponse").Dot("Get").Call(jen.Op("&").Id("result")),
			jen.Return(jen.Id("result"), jen.Err()),
		)).
		// Implement CancelBilling and SetDiscount similarly.
		Add(jen.Func().Params(jen.Id("c").Op("*").Id("customerBillingWorkflowRun")).Id("CancelBilling").Params(
			jen.Id("ctx").Qual("context", "Context"),
			jen.Id("req").Id("CancelBillingSignal"),
			jen.Id("mods").Op("...").Id("WorkflowOptionFunc"),
		).Error().Block(
			jen.Return(
				jen.Id("c").Dot("client").Dot("SignalWorkflow").Call(
					jen.Id("ctx"),
					jen.Id("c").Dot("ID").Call(),
					jen.Id("c").Dot("RunID").Call(),
					jen.Id("barv1CustomerBillingWorkflowCancelBillingName"),
					jen.Id("req"),
				),
			),
		)).
		Add(jen.Func().Params(jen.Id("c").Op("*").Id("customerBillingWorkflowRun")).Id("SetDiscount").Params(
			jen.Id("ctx").Qual("context", "Context"),
			jen.Id("req").Id("SetDiscountSignal"),
			jen.Id("mods").Op("...").Id("WorkflowOptionFunc"),
		).Error().Block(
			jen.Return(
				jen.Id("c").Dot("client").Dot("SignalWorkflow").Call(
					jen.Id("ctx"),
					jen.Id("c").Dot("ID").Call(),
					jen.Id("c").Dot("RunID").Call(),
					jen.Id("barv1CustomerBillingWorkflowSetDiscountName"),
					jen.Id("req"),
				),
			),
		))
}

// GenerateCustomerBillingWorkflowChildRun generates the customerBillingWorkflowChildRun struct and methods.
func GenerateCustomerBillingWorkflowChildRun() *jen.Statement {
	return jen.Type().Id("customerBillingWorkflowChildRun").Struct(
		jen.Id("workflowId").String(),
		jen.Id("runId").String(),
		jen.Id("childFuture").Qual("go.temporal.io/sdk/workflow", "ChildWorkflowFuture"),
	).Line().
		Add(jen.Func().Params(jen.Id("c").Op("*").Id("customerBillingWorkflowChildRun")).Id("ID").Params().String().Block(
			jen.Return(jen.Id("c").Dot("workflowId")),
		)).
		Add(jen.Func().Params(jen.Id("c").Op("*").Id("customerBillingWorkflowChildRun")).Id("IsReady").Params().Bool().Block(
			jen.Return(jen.Id("c").Dot("childFuture").Dot("IsReady").Call()),
		)).
		Add(jen.Func().Params(jen.Id("c").Op("*").Id("customerBillingWorkflowChildRun")).Id("Underlying").Params().Qual("go.temporal.io/sdk/workflow", "ChildWorkflowFuture").Block(
			jen.Return(jen.Id("c").Dot("childFuture")),
		)).
		Add(jen.Func().Params(jen.Id("c").Op("*").Id("customerBillingWorkflowChildRun")).Id("Get").Params(
			jen.Id("ctx").Qual("go.temporal.io/sdk/workflow", "Context"),
		).Params(
			jen.Id("CustomerBillingResponse"),
			jen.Error(),
		).Block(
			jen.Var().Id("result").Id("CustomerBillingResponse"),
			jen.Err().Op(":=").Id("c").Dot("childFuture").Dot("Get").Call(jen.Id("ctx"), jen.Op("&").Id("result")),
			jen.Return(jen.Id("result"), jen.Err()),
		)).
		// Implement CancelBilling and SetDiscount methods similarly.
		Add(jen.Func().Params(jen.Id("c").Op("*").Id("customerBillingWorkflowChildRun")).Id("CancelBilling").Params(
			jen.Id("ctx").Qual("go.temporal.io/sdk/workflow", "Context"),
			jen.Id("req").Id("CancelBillingSignal"),
		).Error().Block(
			jen.Return(
				jen.Id("c").Dot("childFuture").Dot("SignalChildWorkflow").Call(
					jen.Id("ctx"),
					jen.Id("barv1CustomerBillingWorkflowCancelBillingName"),
					jen.Id("req"),
				).Dot("Get").Call(jen.Id("ctx"), jen.Nil()),
			),
		)).
		Add(jen.Func().Params(jen.Id("c").Op("*").Id("customerBillingWorkflowChildRun")).Id("SetDiscount").Params(
			jen.Id("ctx").Qual("go.temporal.io/sdk/workflow", "Context"),
			jen.Id("req").Id("SetDiscountSignal"),
		).Error().Block(
			jen.Return(
				jen.Id("c").Dot("childFuture").Dot("SignalChildWorkflow").Call(
					jen.Id("ctx"),
					jen.Id("barv1CustomerBillingWorkflowSetDiscountName"),
					jen.Id("req"),
				).Dot("Get").Call(jen.Id("ctx"), jen.Nil()),
			),
		))
}
