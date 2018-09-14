package main

import (
	"context"
	"runtime"

	stub "github.com/demo/postgrescontroller/pkg/stub"
	sdk "github.com/operator-framework/operator-sdk/pkg/sdk"
	k8sutil "github.com/operator-framework/operator-sdk/pkg/util/k8sutil"
	sdkVersion "github.com/operator-framework/operator-sdk/version"

	"github.com/sirupsen/logrus"
	"fmt"
)

func printVersion() {
	logrus.Infof("Go Version: %s", runtime.Version())
	logrus.Infof("Go OS/Arch: %s/%s", runtime.GOOS, runtime.GOARCH)
	logrus.Infof("operator-sdk Version: %v", sdkVersion.Version)
}

func main() {
	printVersion()

	sdk.ExposeMetricsPort()

	resource := "postgrescontroller.kubeplus/v1"
	kind := "Postgres"
	namespace, err := k8sutil.GetWatchNamespace()
	if err != nil {
		logrus.Fatalf("Failed to get watch namespace: %v", err)
	}
	resyncPeriod := 0
	logrus.Infof("Watching %s, %s, %s, %d", resource, kind, namespace, resyncPeriod)
	fmt.Println("Hello, I am here.")
	fmt.Println("Whatever you say man")
	sdk.Watch(resource, kind, namespace, resyncPeriod)
	fmt.Println("Do I even make it here?")
	sdk.Handle(stub.NewHandler())
	sdk.Run(context.TODO())
}
