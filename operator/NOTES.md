CR: custom resource eg Org
CRD: Design of Org
Recon: logic that check current state with exepected state. If any diff then do action. ie check 1 instance of indent-consumer is exist or not & create if not
Controller: Watches CR ie Org & run recon on change
Manager: Run multiple Controllers as single process
Client: client package for CRUD k8s objects
Scheme:
Finalizer: 
Owner Reference: 
Webhook:
RBAC:


Files added after adding api

api/v1
|- groupversion_info.go -- contains group version info
|- org_types.go         -- contain CR structure eg : Org, OrgList, OrgSpec & OrgStatus
|- zz_deepcopy.go       -- contain auto_generated code ie deepCopy methods needed for k8s.  updated only by make generate command

internal/controller
|- org_controller.go    -- container reconsiler logic. 
    -- basically it will reconciler struct with following methods
    -- Reconciler struct
    --  |- Reconcile() : WHERE IT HAVING CALL???
    --  |- SetupWithManager() : called in main function to setup it

config/crd
|- kustomization.yanl -- this include path of actual crd.yaml  [but we havent crd.yaml yet. we will create it in next steps]

- update org_types.go as per need for crd object
- run following command
    - task generate :  update zz_generated.deepcopy.go     as per we updated code in org_types.go
    - task manifests : create crd file in config/crd/bases as per we updated code in org_types.go 
                       update config/rbac/role.yaml why? controller need to create crds & update status so we need to add API & permission for such ACTION  
- implements reconcilation loop

packages used:
- corev1:   basic resources ie Pod,Service, ConfigMap, Secrete
- appsv1:   app resources ie deployment & statefulset
- metav1:   metadata resources ie lables
- networkingv1: networking resource ie ingress, network policy
- rbacv1:   rbac resources ie role, rolebinding, clusturerole


NOTE:
- Controller can manage many resources BUT 1 as primary and other as 2ndary
- eg: Deployment is primary resource And service & configMap is secondary resource
- see reconcilation is sticked to primary resource
- BUT when this recon of primary will get trigger?
    1. change in priamry 
    2. change in secondary resources also trigger reconcilation of its parent/primary resource
- What If I want to do recon on resource other than primary/2ndary ie purely independant -> We can do with Watches function
    But again we have 3 options
        - Reconcile the changed object
        - Reconcile the owner of changed object
        - Reconcile with any resource with custom logic
    this will be done by passing 2nd arg ti Watch function

eg:
    ctrl.NewControllerManagedBy(manager). -- return controller that manage by manager
		For(&appsv1.Deployment{}).  -- define primary resource for controller
		Owns(&corev1.Service{}).    -- define 2ndary resource
        Owns(&corev1.ConfigMap{}).    -- define 2ndary resource
		Watches(&corev1.Secret{}, &handler.EnqueueRequestForObject{}). -- define independant resource you willing to watch even its not part of primary & secondary resource. 2nd arg can define reconcilaion whome ie own/parent/custom
		Complete(r)

Create manager:
|- accpet two ars
        1. Kubeconfig
        2. Scheme -- what is scheme here
    
Scheme:
|- WHAT DOES : do mapping of k8s go structs => respective api_groups
|- WHAT IT IS : type registry ie register type 
|- WHY NEEDED : client need to know which api execute for this struct

So always create new empty scheme, call AddToScheme(scheme) which will register all k8s buildin resource in that scheme & now use this scheme as arg to manager. 
Now how to register CR in schema??  -> our CR should have its own AddToScheme() which we will call after registering buildin resources.
