Problem:
---------
Set up a Postgres instance per customer on Kubernetes adhering to following requirements:

1) The Postgres instance should be initialized with specified databases and users.
It should be optionally possible to specify creation and initialization of 
tables on one or more databases.

2) It should be possible to perform following operations on an already created Postgres instance.
- modify a particular user's password
- create new databases
- create new users
- delete existing users


Solution:
----------
Below we present five possible approaches towards addressing the Postgres setup problem.


1) Building custom Postgres images:
   --------------------------------
   This approach consists of building a custom Postgres image using a customer-specific
   Dockerfile. Postgres supports creation of custom database, 
   user, and password by defining appropriate environment variables.
   Define these variables in the Dockerfile.

   See per-customer-image directory for details.

   Advantage:
   - Easiest approach to get started.

   Disadvantages:
   - A custom container needs to be built for each customer.
   - Separate automation needs to be written for performing database actions later (such as modify user password,
     add new users, add new databases, etc.). This automation consists of 
     remotely connecting to the running container (its Service/Deployment), and performing
     the required actions.


2) Use postStart container life-cycle hook:
   ----------------------------------------
   This approach consists of using a postStart container lifecycle hook. We create a custom script 
   that creates databases and users using the data that is passed to this script at runtime.
   We build a single custom container that contains this script.
   Then for each customer, we generate the required data (database names, usernames, passwords),
   and pass it to the postStart hook script through environment variables.
   Example - https://stackoverflow.com/questions/48150137/can-i-use-env-in-poststart-command

   Advantage:
   - Container is built only once.

   Disadvantages:
   - Custom container still needs to be built.
   - Separate automation needs to be written for performing database actions later (such as modify user password,
     add new users, add new databases, etc.). This automation consists of 
     remotely connecting to the running container (its Service/Deployment), and performing
     the required actions.


3) Out-of-band orchestration:
   ---------------------------
   This approach consists of creating a Postgres Deployment/StatefulSet using a standard Postgres image.
   The deployment is exposed through a Kubernetes Service.
   Once the Service is READY, a script is run against this Service IP remotely which initializes the
   Postgres instance.

   See out-of-band directory for details.

   Advantage:
   - Easy to integrate with existing automation - if there already exists some automation.

   Disadvantages:
   - Considerable state needs to be maintained on your end to track the Service and to execute
     database creation script against it.
   - Additional custom scripting is required to perform database actions later on.


4) Helm charts with post-install chart hook:
   -----------------------------------------
   This approach consists of using the Postgres Helm chart [5] and enhancing it with
   post-install hook that will setup the Postgres instance using steps similar to approach 2.

   Advantage:
   - 
   
   Disadvantages:
   - Same as approach 2.
   - Infact, approach 2 is bit easier than this approach (no understanding of Helm is required in
     that approach).


5) Postgres Operator / Custom resource controller:
   -----------------------------------------------
   Use a Postgres Operator / Custom Resource Controller (CRD) such as [1, 2, 3, 4].

   Advantages (of [1]):
   - Declarative inputs - Required databases and users are specified declaratively
     in the Spec of the CRD.
   - Tables and data initialization - The CRD Spec supports ability to specify initialization commands
     such as creating and populating tables with required data.
   - Declarative Updates - Performing updates to a Postgres instance is straightforward. 
     Updating a instance amounts to updating the required declarative attributes
     in the CRD YAML with new data and then re-applying it using kubectl.
   - No out-of-band custom scripting needed.
   - Kubernetes-native - All the database setup and modification actions are done using 'kubectl'.
     No need to use other CLI.

   Disadvantage:
   - Postgres Operator needs to be written if an existing one does not satisfy your needs.


References:
------------
[1] https://github.com/cloud-ark/kubeplus/tree/master/postgres-crd-v2
[2] https://github.com/kubedb/postgres
[3] https://github.com/CrunchyData/postgres-operator
[4] https://github.com/zalando-incubator/postgres-operator
[5] https://github.com/kubernetes/charts/tree/master/stable/postgresql

