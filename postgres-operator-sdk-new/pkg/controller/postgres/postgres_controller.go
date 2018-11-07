package postgres

import (
	postgresv1 "labs/postgres-operator-sdk-new/pkg/apis/postgresoperatorsdknew/v1"
	"log"
	"os"
	"io/ioutil"
	"context"
	"strconv"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"labs/postgres-operator-sdk-new/pkg/apis/postgresoperatorsdknew/v1"
	"k8s.io/apimachinery/pkg/types"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/controller"
	"sigs.k8s.io/controller-runtime/pkg/handler"
	"sigs.k8s.io/controller-runtime/pkg/manager"
	"sigs.k8s.io/controller-runtime/pkg/reconcile"
	"sigs.k8s.io/controller-runtime/pkg/source"
	appsv1 "k8s.io/api/apps/v1"
	apiv1 "k8s.io/api/core/v1"
	apiutil "k8s.io/apimachinery/pkg/util/intstr"
	"k8s.io/apimachinery/pkg/runtime/schema"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	"strings"
	"fmt"
	"time"

	_ "github.com/lib/pq"
	"database/sql"
)

const controllerAgentName = "postgres-controller"

const (
	// SuccessSynced is used as part of the Event 'reason' when a Foo is synced
	SuccessSynced = "Synced"
	// ErrResourceExists is used as part of the Event 'reason' when a Foo fails
	// to sync due to a Deployment of the same name already existing.
	ErrResourceExists = "ErrResourceExists"

	// MessageResourceExists is the message used for Events when a resource
	// fails to sync due to a Deployment already existing
	MessageResourceExists = "Resource %q already exists and is not managed by Foo"
	// MessageResourceSynced is the message used for an Event fired when a Foo
	// is synced successfully
	MessageResourceSynced = "Foo synced successfully"
)

const (
      PGPASSWORD = "mysecretpassword"
      MINIKUBE_IP = "192.168.99.100"
)

var (
	id_list = []string{}
)


/**
* USER ACTION REQUIRED: This is a scaffold file intended for the user to modify with their own Controller
* business logic.  Delete these comments after modifying this file.*
 */

// Add creates a new Postgres Controller and adds it to the Manager. The Manager will set fields on the Controller
// and Start it when the Manager is Started.
func Add(mgr manager.Manager) error {
	return add(mgr, newReconciler(mgr))
}

// newReconciler returns a new reconcile.Reconciler
func newReconciler(mgr manager.Manager) reconcile.Reconciler {
	return &ReconcilePostgres{client: mgr.GetClient(), scheme: mgr.GetScheme()}
}

// add adds a new Controller to mgr with r as the reconcile.Reconciler
func add(mgr manager.Manager, r reconcile.Reconciler) error {
	// Create a new controller
	log.Print("I AM HERE\n")
	c, err := controller.New("postgres-controller", mgr, controller.Options{Reconciler: r})
	if err != nil {
		return err
	}

	log.Print("Did I make it here????\n")

	// Watch for changes to primary resource Postgres
	err = c.Watch(&source.Kind{Type: &v1.Postgres{}}, &handler.EnqueueRequestForObject{})
	if err != nil {
		return err
	}

	log.Print("I also made it here!!!\n")
	// TODO(user): Modify this to be the types you create that are owned by the primary resource
	// Watch for changes to secondary resource Pods and requeue the owner Postgres
	err = c.Watch(&source.Kind{Type: &corev1.Pod{}}, &handler.EnqueueRequestForOwner{
		IsController: true,
		OwnerType:    &postgresv1.Postgres{},
	})
	if err != nil {
		return err
	}

	return nil
}

var _ reconcile.Reconciler = &ReconcilePostgres{}

// ReconcilePostgres reconciles a Postgres object
type ReconcilePostgres struct {
	// This client, initialized using mgr.Client() above, is a split client
	// that reads objects from the cache and writes to the apiserver
	client client.Client
	scheme *runtime.Scheme
}

// Reconcile reads that state of the cluster for a Postgres object and makes changes based on the state read
// and what is in the Postgres.Spec
// TODO(user): Modify this Reconcile function to implement your Controller logic.  This example creates
// a Pod as an example
// Note:
// The Controller will requeue the Request to be processed again if the returned error is non-nil or
// Result.Requeue is true, otherwise upon completion it will remove the work from the queue.
func (r *ReconcilePostgres) Reconcile(request reconcile.Request) (reconcile.Result, error) {
	log.Printf("Reconciling Postgres %s/%s\n", request.Namespace, request.Name)

	// Fetch the Postgres instance
	instance := &postgresv1.Postgres{}
	err := r.client.Get(context.TODO(), request.NamespacedName, instance)
	if err != nil {
		if errors.IsNotFound(err) {
			// Request object not found, could have been deleted after reconcile request.
			// Owned objects are automatically garbage collected. For additional cleanup logic use finalizers.
			// Return and don't requeue
			return reconcile.Result{}, nil
		}
		// Error reading the object - requeue the request.
		return reconcile.Result{}, err
	}

	deploymentName := instance.Spec.DeploymentName

	var verifyCmd string
	var actionHistory []string
	var serviceIP string
	var servicePort string
	var setupCommands []string
	var databases []string
	var users []postgresv1.UserSpec

	serviceIP, servicePort, setupCommands, databases, users, verifyCmd, err = createDeployment(instance, r)

	if err != nil {
		if apierrors.IsAlreadyExists(err) {

			fmt.Printf("CRD %s created\n", deploymentName)
			fmt.Printf("Check using: kubectl describe postgres %s \n", deploymentName)

			pgresObj := instance
			err2 := r.client.Get(context.TODO(), request.NamespacedName, pgresObj)

			if err2 != nil {
				panic(err2)
			}

			actionHistory := pgresObj.Status.ActionHistory
			serviceIP := pgresObj.Status.ServiceIP
			servicePort := pgresObj.Status.ServicePort
			verifyCmd := pgresObj.Status.VerifyCmd
			fmt.Printf("Action History:[%s]\n", actionHistory)
			fmt.Printf("Service IP:[%s]\n", serviceIP)
			fmt.Printf("Service Port:[%s]\n", servicePort)
			fmt.Printf("Verify cmd: %v\n", verifyCmd)

			//setupCommands = canonicalize(foo.Spec.Commands)

			var commandsToRun []string

			desiredDatabases := instance.Spec.Databases
			currentDatabases := pgresObj.Status.Databases
			fmt.Printf("Current Databases:%v\n", currentDatabases)
			fmt.Printf("Desired Databases:%v\n", desiredDatabases)
			createDBCommands, dropDBCommands := getDatabaseCommands(desiredDatabases,
				currentDatabases)
			appendList(&commandsToRun, createDBCommands)
			appendList(&commandsToRun, dropDBCommands)

			desiredUsers := instance.Spec.Users
			currentUsers := pgresObj.Status.Users
			fmt.Printf("Current Users:%v\n", currentUsers)
			fmt.Printf("Desired Users:%v\n", desiredUsers)
			createUserCmds, dropUserCmds, alterUserCmds := getUserCommands(desiredUsers,
				currentUsers)
			appendList(&commandsToRun, createUserCmds)
			appendList(&commandsToRun, dropUserCmds)
			appendList(&commandsToRun, alterUserCmds)

			//commandsToRun = getCommandsToRun(actionHistory, setupCommands)
			fmt.Printf("commandsToRun: %v\n", commandsToRun)

			if len(commandsToRun) > 0 {
				err2 := updateFooStatus(instance, &actionHistory, &currentUsers, &desiredDatabases,
					verifyCmd, serviceIP, servicePort, "UPDATING", r)
				if err2 != nil {
					return reconcile.Result{}, err2
				}
				updateCRD(pgresObj, commandsToRun)
			}

			pgresObj2 := &postgresv1.Postgres{}
			err3 := r.client.Get(context.TODO(), request.NamespacedName, pgresObj2)

			if err3 != nil {
				panic(err3)
			}

			actionHistory = pgresObj2.Status.ActionHistory
			fmt.Printf("1111 Action History:%s\n", actionHistory)
			for _, cmds := range commandsToRun {
				actionHistory = append(actionHistory, cmds)
			}

			err3 = updateFooStatus(pgresObj2, &actionHistory, &desiredUsers, &desiredDatabases,
				verifyCmd, serviceIP, servicePort, "READY", r)

			if err3 != nil {
				panic(err)
				return reconcile.Result{}, err3
			}

		} else {
			panic(err)
		}
	} else {
		for _, cmds := range setupCommands {
			if !strings.Contains(cmds, "\\c") {
				actionHistory = append(actionHistory, cmds)
			}
		}
		fmt.Printf("Setup Commands: %v\n", setupCommands)
		fmt.Printf("Verify using: %v\n", verifyCmd)

		err1 := updateFooStatus(instance, &actionHistory, &users, &databases, verifyCmd, serviceIP, servicePort, "READY", r)

		if err1 != nil {
			return reconcile.Result{}, err1
		}
	}

	return reconcile.Result{}, nil
}

func updateCRD(foo *postgresv1.Postgres, setupCommands []string) {
	serviceIP := foo.Status.ServiceIP
	servicePort := foo.Status.ServicePort

	fmt.Printf("Service IP:[%s]\n", serviceIP)
	fmt.Printf("Service Port:[%s]\n", servicePort)
	fmt.Printf("Command:[%s]\n", setupCommands)

	if len(setupCommands) > 0 {
		//file := createTempDBFile(setupCommands)
		fmt.Println("Now setting up the database")
		//setupDatabase(serviceIP, servicePort, file)
		var dummyList []string
		setupDatabase(serviceIP, servicePort, setupCommands, dummyList)
	}
}

func updateFooStatus(foo *postgresv1.Postgres,
	actionHistory *[]string, users *[]postgresv1.UserSpec, databases *[]string,
	verifyCmd string, serviceIP string, servicePort string,
	status string, r *ReconcilePostgres) error {

	fooCopy := foo.DeepCopy()
	fooCopy.Status.AvailableReplicas = 1

	fooCopy.Status.VerifyCmd = verifyCmd
	fooCopy.Status.ActionHistory = *actionHistory
	fooCopy.Status.Users = *users
	fooCopy.Status.Databases = *databases
	fooCopy.Status.ServiceIP = serviceIP
	fooCopy.Status.ServicePort = servicePort
	fooCopy.Status.Status = status
	err := r.client.Update(context.TODO(), fooCopy)
	return err
}

func createDeployment (foo *postgresv1.Postgres, r *ReconcilePostgres) (string, string, []string, []string, []postgresv1.UserSpec, string, error){

	deploymentName := foo.Spec.DeploymentName
	image := foo.Spec.Image
	users := foo.Spec.Users
	databases := foo.Spec.Databases
	setupCommands := canonicalize(foo.Spec.Commands)

	var userAndDBCommands []string
	var allCommands []string

	var currentDatabases []string
	var currentUsers []postgresv1.UserSpec
	createDBCmds, dropDBCmds := getDatabaseCommands(databases, currentDatabases)
	createUserCmds, dropUserCmds, alterUserCmds := getUserCommands(users, currentUsers)

	fmt.Printf("   Deployment:%v, Image:%v\n", deploymentName, image)
	fmt.Printf("   Users:%v\n", users)
	fmt.Printf("   Databases:%v\n", databases)
	fmt.Printf("   SetupCmds:%v\n", setupCommands)
	fmt.Printf("   CreateDBCmds:%v\n", createDBCmds)
	fmt.Printf("   DropDBCmds:%v\n", dropDBCmds)
	fmt.Printf("   CreateUserCmds:%v\n", createUserCmds)
	fmt.Printf("   DropUserCmds:%v\n", dropUserCmds)
	fmt.Printf("   AlterUserCmds:%v\n", alterUserCmds)

	appendList(&userAndDBCommands, createDBCmds)
	appendList(&userAndDBCommands, dropDBCmds)
	appendList(&userAndDBCommands, createUserCmds)
	appendList(&userAndDBCommands, dropUserCmds)
	appendList(&userAndDBCommands, alterUserCmds)
	fmt.Printf("   UserAndDBCmds:%v\n", userAndDBCommands)
	fmt.Printf("   SetupCmds:%v\n", setupCommands)

	appendList(&allCommands, userAndDBCommands)
	appendList(&allCommands, setupCommands)

	deployment := &appsv1.Deployment{
		TypeMeta: metav1.TypeMeta{
			APIVersion: "apps/v1",
			Kind:       "Deployment",
		},
		ObjectMeta: metav1.ObjectMeta{
			Name: deploymentName,
			Namespace: "default",
		},
		Spec: appsv1.DeploymentSpec{
			Replicas: int32Ptr(1),
			Selector: &metav1.LabelSelector{
				MatchLabels: map[string]string{
					"app": deploymentName,
				},
			},
			Template: apiv1.PodTemplateSpec{
				ObjectMeta: metav1.ObjectMeta{
					Labels: map[string]string{
						"app": deploymentName,
					},
				},

				Spec: apiv1.PodSpec{
					Containers: []apiv1.Container{
						{
							Name:  deploymentName,
							Image: image,
							Ports: []apiv1.ContainerPort{
								{
									ContainerPort: 5432,
								},
							},
							ReadinessProbe: &apiv1.Probe{
								Handler: apiv1.Handler{
									TCPSocket: &apiv1.TCPSocketAction{
										Port: apiutil.FromInt(5432),
									},
								},
								InitialDelaySeconds: 5,
								TimeoutSeconds: 60,
								PeriodSeconds: 2,
							},
							Env: []apiv1.EnvVar{
								{
									Name: "POSTGRES_PASSWORD",
									Value: PGPASSWORD,
								},
							},
						},
					},
				},
			},
		},
	}

	fmt.Println("Creating deployment...")
	err := r.client.Create(context.TODO(), deployment)
	fmt.Println(err)
	if err != nil {
		return "", "", nil, nil, nil, "", err
	}

	fmt.Printf("Created deployment %q.\n", deployment.GetObjectMeta().GetName())
	fmt.Printf("------------------------------\n")

	fmt.Printf("Creating service.....\n")

	service := &apiv1.Service{
		TypeMeta: metav1.TypeMeta{
			APIVersion: "v1",
			Kind:       "Service",
		},
		ObjectMeta: metav1.ObjectMeta{
			Name: deploymentName,
			Namespace: "default",
			Labels: map[string]string{
				"app": deploymentName,
			},
		},
		Spec: apiv1.ServiceSpec{
			Ports: []apiv1.ServicePort {
				{
					Name: "my-port",
					Port: 5432,
					TargetPort: apiutil.FromInt(5432),
					Protocol: apiv1.ProtocolTCP,
				},
			},
			Selector: map[string]string {
				"app": deploymentName,
			},
			Type: apiv1.ServiceTypeNodePort,
		},
	}

	found := &apiv1.Service{}
	err2 := r.client.Get(context.TODO(), types.NamespacedName{Name: deploymentName, Namespace: "default"}, found)
	if err2 != nil {
		fmt.Println(err2)
	}
	
	err1 := r.client.Create(context.TODO(), service)

	if err1 != nil {
		return "", "", nil, nil, nil,  "", err
	}

	fmt.Printf("Created service %q.\n", service.GetObjectMeta().GetName())
	fmt.Printf("------------------------------\n")

	serviceIP := MINIKUBE_IP

	nodePort1 := service.Spec.Ports[0].NodePort
	nodePort := fmt.Sprint(nodePort1)
	servicePort := nodePort

	time.Sleep(time.Second * 5)

	for {
		readyPods := 0
		pods := getPods()

		for _, d := range pods.Items {
			podConditions := d.Status.Conditions
			for _, podCond := range podConditions {
				if podCond.Type == corev1.PodReady {
					if podCond.Status == corev1.ConditionTrue {
						readyPods += 1
					}
				}
			}
		}

		if readyPods >= len(pods.Items) {
			break
		} else {
			fmt.Println("Waiting for Pod to get ready.")
			time.Sleep(time.Second * 4)
		}
	}

	time.Sleep(time.Second * 2)

	if len(userAndDBCommands) > 0 {
		fmt.Printf("About to create temp db file for user and db commands")
		//file := createTempDBFile(userAndDBCommands)
		fmt.Println("Now setting up the database")
		//setupDatabase_prev(serviceIP, servicePort, file)
		var dummyList []string
		setupDatabase(serviceIP, servicePort, userAndDBCommands, dummyList)
	}

	if len(setupCommands) > 0 {
		fmt.Printf("About to create temp db file for setup commands")
		//file := createTempDBFile(setupCommands)
		fmt.Println("Now setting up the database")
		//setupDatabase(serviceIP, servicePort, file)
		setupDatabase(serviceIP, servicePort, setupCommands, databases)
	}

	verifyCmd := strings.Fields("psql -h " + serviceIP + " -p " + nodePort + " -U <user> " + " -d <db-name>")
	var verifyCmdString = strings.Join(verifyCmd, " ")
	fmt.Printf("VerifyCmd: %v\n", verifyCmd)
	return serviceIP, servicePort, allCommands, databases, users, verifyCmdString, err

}

func setupDatabase(serviceIP string, servicePort string, setupCommands []string, databases []string) {

	fmt.Println("Setting up database")
	fmt.Println("Commands:")
	fmt.Printf("%v", setupCommands)

	var host = serviceIP
	port := -1
	port, _ = strconv.Atoi(servicePort)
	var user = "postgres"
	var password = PGPASSWORD

	var psqlInfo string
	if len(databases) > 0 {
		dbname := databases[0]
		fmt.Println("%s\n", dbname)
		psqlInfo = fmt.Sprintf("host=%s port=%d user=%s "+
			"password=%s dbname=%s sslmode=disable",
			host, port, user, password, dbname)
	} else {
		psqlInfo = fmt.Sprintf("host=%s port=%d user=%s "+
			"password=%s sslmode=disable",
			host, port, user, password)
	}

	db, err := sql.Open("postgres", psqlInfo)
	if err != nil {
		panic(err)
	}
	defer db.Close()

	err = db.Ping()
	if err != nil {
		panic(err)
	}

	fmt.Println("Successfully connected!")

	for _, command := range setupCommands {
		_, err = db.Exec(command)
		if err != nil {
			panic(err)
		}
	}
	fmt.Println("Done setting up the database")
}

func createTempDBFile(setupCommands []string) (*os.File){
	file, err := ioutil.TempFile("/tmp", "create-db1")
	if err != nil {
		panic(err)
	}

	fmt.Printf("Database setup file:%s\n", file.Name())

	for _, command := range setupCommands {
		//fmt.Printf("Command: %v\n", command)
		// TODO: Interpolation of variables
		file.WriteString(command)
		file.WriteString("\n")
	}
	file.Sync()
	file.Close()
	return file
}

// podList returns a v1.PodList object
func getPods() *apiv1.PodList {
	return &apiv1.PodList{
		TypeMeta: metav1.TypeMeta{
			Kind:       "Pod",
			APIVersion: "v1",
		},
	}
}

func int32Ptr(i int32) *int32 { return &i }

// newbusyBoxPod demonstrates how to create a busybox pod
func newbusyBoxPod(cr *postgresv1.Postgres) *corev1.Pod {
	labels := map[string]string{
		"app": "busy-box",
	}
	return &corev1.Pod{
		TypeMeta: metav1.TypeMeta{
			Kind:       "Pod",
			APIVersion: "v1",
		},
		ObjectMeta: metav1.ObjectMeta{
			Name:      "busy-box",
			Namespace: cr.Namespace,
			OwnerReferences: []metav1.OwnerReference{
				*metav1.NewControllerRef(cr, schema.GroupVersionKind{
					Group:   v1.SchemeGroupVersion.Group,
					Version: v1.SchemeGroupVersion.Version,
					Kind:    "Postgres",
				}),
			},
			Labels: labels,
		},
		Spec: corev1.PodSpec{
			Containers: []corev1.Container{
				{
					Name:    "busybox",
					Image:   "busybox",
					Command: []string{"sleep", "3600"},
				},
			},
		},
	}
}


