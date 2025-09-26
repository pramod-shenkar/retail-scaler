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
