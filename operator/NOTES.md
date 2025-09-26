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
