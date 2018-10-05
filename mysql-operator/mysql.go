package main

import (
        "fmt"
		"strings"
		operatorv1 "github.com/cloud-ark/moodle-operator/pkg/apis/moodlecontroller/v1"
		appsv1 "k8s.io/api/apps/v1"
		metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
		apiutil "k8s.io/apimachinery/pkg/util/intstr"
		apiv1 "k8s.io/api/core/v1"
		"k8s.io/apimachinery/pkg/api/resource"
		"database/sql"
		_ "github.com/go-sql-driver/mysql"
		"time"
		//"strconv"
		corev1 "k8s.io/api/core/v1"
)

var (
	MYSQL_ROOT_PASSWORD  = "mysecretpassword"
	MYSQL_PORT = 3306
	API_VERSION = "moodlecontroller.cloudark/v1"
	MOODLE_KIND = "Moodle"
)

func (c *Controller) deployMysql(foo *operatorv1.Moodle) (string, string) {
	fmt.Println("Inside deployMysql")
	c.createPersistentVolume(foo)
	c.createPersistentVolumeClaim(foo)
	c.createDeployment(foo)
	serviceIP, servicePort := c.createService(foo)

	// Wait couple of seconds more just to give the Pod some more time.
	time.Sleep(time.Second * 2)

	c.setupDatabase(serviceIP, servicePort)

	verifyCmd := strings.Fields("mysql --host=" + serviceIP + " --port=" + servicePort + " --user=root" + " --password=mysecretpassword")
	var verifyCmdString = strings.Join(verifyCmd, " ")
	fmt.Printf("VerifyCmd: %v\n", verifyCmd)
	serviceIPToReturn := serviceIP + ":" + servicePort
	return serviceIPToReturn, verifyCmdString
}

func (c *Controller) setupDatabase(serviceIP, servicePort string) {
	fmt.Println("Setting up database")
	var setupCommands []string
	setupCommands = make([]string, 0)
	setupCommands = append(setupCommands, "CREATE DATABASE moodle DEFAULT CHARACTER SET utf8 COLLATE utf8_unicode_ci")

	//dbHostString := serviceIP + ":" + servicePort
	dbHostString := serviceIP
	passwordString := " IDENTIFIED BY 'passwordformoodleadmin'"
	createUserCmd1 := "create user 'moodleadmin'@'" + dbHostString + "'" + passwordString
	createUserCmd2 := "create user 'moodleadmin'@'%'" + passwordString
	setupCommands = append(setupCommands, createUserCmd1)
	setupCommands = append(setupCommands, createUserCmd2)

	grantCmd1 := "GRANT SELECT,INSERT,UPDATE,DELETE,CREATE,CREATE TEMPORARY TABLES,DROP,INDEX,ALTER ON moodle.* TO "
	grantCmd1 = grantCmd1 + "'moodleadmin'@'" + dbHostString + "'" + passwordString
	setupCommands = append(setupCommands, grantCmd1)

	grantCmd2 := "GRANT SELECT,INSERT,UPDATE,DELETE,CREATE,CREATE TEMPORARY TABLES,DROP,INDEX,ALTER ON moodle.* TO "
	grantCmd2 = grantCmd2 + "moodleadmin@'%'" + passwordString
	setupCommands = append(setupCommands, grantCmd2)

	//setupCommands = append(setupCommands, "create user 'moodleadmin'@'localhost' IDENTIFIED BY 'passwordformoodleadmin'")
	//setupCommands = append(setupCommands, "GRANT SELECT,INSERT,UPDATE,DELETE,CREATE,CREATE TEMPORARY TABLES,DROP,INDEX,ALTER ON moodle.* TO moodleadmin@localhost IDENTIFIED BY 'passwordformoodleadmin'")

	fmt.Println("Commands:")
	fmt.Printf("%v", setupCommands)

	//var host = serviceIP
	//port := -1
	//port, _ = strconv.Atoi(servicePort)
	var user = "root"
	var password = MYSQL_ROOT_PASSWORD

	//var mysqlInfo string
	//mysqlInfo = fmt.Sprintf("%s:%s@tcp(%s:%s)", user, password, host, port)

	//user:password@tcp(127.0.0.1:3306)

	//dsn := user + ":" + password + "@/" + dbname

	dsn := user + ":" + password + "@tcp(" + serviceIP + ":" + servicePort + ")/"
	fmt.Printf("DSN:%s\n", dsn)

	db, err := sql.Open("mysql", dsn)
	if err != nil {
		panic(err)
	}
	defer db.Close()

	err = db.Ping()
	if err != nil {
		panic(err)
	}

	for _, command := range setupCommands {
		_, err = db.Exec(command)
		if err != nil {
			panic(err)
		}
	}
	fmt.Println("Done setting up the database")
}

func (c *Controller) createPersistentVolume(foo *operatorv1.Moodle) {
	fmt.Println("Inside createPersistentVolume")

	deploymentName := foo.Spec.Name
	persistentVolume := &apiv1.PersistentVolume{
			ObjectMeta: metav1.ObjectMeta{
				Name: deploymentName,
				OwnerReferences: []metav1.OwnerReference{
				{
					APIVersion: API_VERSION,
					Kind: MOODLE_KIND,
					Name: foo.Name,
					UID: foo.UID,
				},
			},
		},
		Spec: apiv1.PersistentVolumeSpec{
				StorageClassName: "manual",
				Capacity: apiv1.ResourceList{
//					map[string]resource.Quantity{
						"storage": resource.MustParse("1Gi"),
//					},
				},
				AccessModes: []apiv1.PersistentVolumeAccessMode{
//					{
						"ReadWriteOnce",
//					},
				},
				PersistentVolumeSource: apiv1.PersistentVolumeSource{
					HostPath: &apiv1.HostPathVolumeSource{
						Path: "/mnt/mysql-data",
					},
				},
			},
	}

	persistentVolumeClient := c.kubeclientset.CoreV1().PersistentVolumes()

	fmt.Println("Creating persistentVolume...")
	result, err := persistentVolumeClient.Create(persistentVolume)
	if err != nil {
		panic(err)
	}
	fmt.Printf("Created persistentVolume %q.\n", result.GetObjectMeta().GetName())

}

func (c *Controller) createPersistentVolumeClaim(foo *operatorv1.Moodle) {
	fmt.Println("Inside createPersistentVolumeClaim")

	deploymentName := foo.Spec.Name
	persistentVolumeClaim := &apiv1.PersistentVolumeClaim{
			ObjectMeta: metav1.ObjectMeta{
				Name: deploymentName,
				OwnerReferences: []metav1.OwnerReference{
				{
					APIVersion: API_VERSION,
					Kind: MOODLE_KIND,
					Name: foo.Name,
					UID: foo.UID,
				},
			},
		},
		Spec: apiv1.PersistentVolumeClaimSpec{
				AccessModes: []apiv1.PersistentVolumeAccessMode{
//					{
						"ReadWriteOnce",
//					},
				},
				Resources: apiv1.ResourceRequirements{
					Requests: apiv1.ResourceList {
							"storage": resource.MustParse("1Gi"),
//							map[string]resource.Quantity{
//							"storage": resource.MustParse("1Gi"),
//						},
					},
				},
		},
	}

	persistentVolumeClaimClient := c.kubeclientset.CoreV1().PersistentVolumeClaims(apiv1.NamespaceDefault)

	fmt.Println("Creating persistentVolumeClaim...")
	result, err := persistentVolumeClaimClient.Create(persistentVolumeClaim)
	if err != nil {
		panic(err)
	}
	fmt.Printf("Created persistentVolumeClaim %q.\n", result.GetObjectMeta().GetName())
}


func (c *Controller) createDeployment(foo *operatorv1.Moodle) {

	fmt.Println("Inside createDeployment")

	deploymentsClient := c.kubeclientset.AppsV1().Deployments(apiv1.NamespaceDefault)

	deploymentName := foo.Spec.Name
	image := "mysql:5.6"
	volumeName := "mysql-persistent-storage"

	deployment := &appsv1.Deployment{
		ObjectMeta: metav1.ObjectMeta{
			Name: deploymentName,
			OwnerReferences: []metav1.OwnerReference{
				{
					APIVersion: API_VERSION,
					Kind: MOODLE_KIND,
					Name: foo.Name,
					UID: foo.UID,
				},
			},
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
									ContainerPort: 3306,
								},
							},
							ReadinessProbe: &apiv1.Probe{
								Handler: apiv1.Handler{
									TCPSocket: &apiv1.TCPSocketAction{
										Port: apiutil.FromInt(MYSQL_PORT),
									},
								},
								InitialDelaySeconds: 5,
								TimeoutSeconds:      60,
								PeriodSeconds:       2,
							},
							Env: []apiv1.EnvVar{
								{
									Name:  "MYSQL_ROOT_PASSWORD",
									Value: MYSQL_ROOT_PASSWORD,
								},
							},
							VolumeMounts: []apiv1.VolumeMount{
								{
									Name: volumeName,
									MountPath: "/var/lib/mysql",

								},
							},	
						},
					},
					Volumes: []apiv1.Volume{
						{
							Name: volumeName,
							VolumeSource: apiv1.VolumeSource{
								PersistentVolumeClaim: &apiv1.PersistentVolumeClaimVolumeSource{
										ClaimName: deploymentName,
								},
							},
						},
					},
				},
			},
		},
	}

	// Create Deployment
	fmt.Println("Creating deployment...")
	result, err := deploymentsClient.Create(deployment)
	if err != nil {
		panic(err)
	}
	fmt.Printf("Created deployment %q.\n", result.GetObjectMeta().GetName())
}


func (c *Controller) createService(foo *operatorv1.Moodle) (string, string) {

	fmt.Println("Inside createService")
	deploymentName := foo.Spec.Name

	serviceClient := c.kubeclientset.CoreV1().Services(apiv1.NamespaceDefault)
	service := &apiv1.Service{
		ObjectMeta: metav1.ObjectMeta{
			Name: deploymentName,
			OwnerReferences: []metav1.OwnerReference{
				{
					APIVersion: API_VERSION,
					Kind: MOODLE_KIND,
					Name: foo.Name,
					UID: foo.UID,
				},
			},
			Labels: map[string]string{
				"app": deploymentName,
			},
		},
		Spec: apiv1.ServiceSpec{
			Ports: []apiv1.ServicePort{
				{
					Name:       "my-port",
					Port:       3306,
					TargetPort: apiutil.FromInt(3306),
					Protocol:   apiv1.ProtocolTCP,
				},
			},
			Selector: map[string]string{
				"app": deploymentName,
			},
			Type: apiv1.ServiceTypeNodePort,
		},
	}

	result1, err1 := serviceClient.Create(service)
	if err1 != nil {
		panic(err1)
	}
	fmt.Printf("Created service %q.\n", result1.GetObjectMeta().GetName())

	// Parse ServiceIP and Port
	serviceIP := HOST_IP
	fmt.Println("HOST IP:%s", serviceIP)

	nodePort1 := result1.Spec.Ports[0].NodePort
	nodePort := fmt.Sprint(nodePort1)
	servicePort := nodePort

	c.waitForPod(foo)

	serviceIPToReturn := serviceIP + ":" + servicePort

	fmt.Printf("Service IP to Return:%s\n", serviceIPToReturn)

	return serviceIP, servicePort
}

func (c *Controller) waitForPod(foo *operatorv1.Moodle) {

	deploymentName := foo.Spec.Name
	// Check if Postgres Pod is ready or not
	podReady := false
	for {
		pods := c.getPods(deploymentName)
		for _, d := range pods.Items {
			if strings.Contains(d.Name, deploymentName) {
				podConditions := d.Status.Conditions
				for _, podCond := range podConditions {
					if podCond.Type == corev1.PodReady {
						if podCond.Status == corev1.ConditionTrue {
							fmt.Println("MySQL Pod is running.")
							podReady = true
							break
						}
					}
				}
			}
			if podReady {
				break
			}
		}
		if podReady {
			break
		} else {
			fmt.Println("Waiting for MySQL Pod to get ready.")
			time.Sleep(time.Second * 4)
		}
	}
	fmt.Println("Pod is ready.")
}

func (c *Controller) getPods(deploymentName string) *apiv1.PodList {
	// TODO(devkulkarni): This is returning all Pods. We should change this
	// to only return Pods whose Label matches the Deployment Name.
	pods, err := c.kubeclientset.CoreV1().Pods("default").List(metav1.ListOptions{
		//LabelSelector: deploymentName,
		//LabelSelector: metav1.LabelSelector{
		//	MatchLabels: map[string]string{
		//	"app": deploymentName,
		//},
		//},
	})
	//fmt.Printf("There are %d pods in the cluster\n", len(pods.Items))
	if err != nil {
		fmt.Printf("%s", err)
	}
	return pods
}

func int32Ptr(i int32) *int32 { return &i }