## This argokit v2:

local argokit = import '../../argokit/v2/jsonnet/argokit.libsonnet';
local application = argokit.appAndObjects.application;

application.new(
  name='foo-frontend',
  image='foo-frontend:1.2.3',
  port=3000
)

## Renders to this Kubernetes Application manifest:

{
   "apiVersion": "v1",
   "items": [
      {
         "apiVersion": "skiperator.kartverket.no/v1alpha1",
         "kind": "Application",
         "metadata": {
            "name": "foo-frontend"
         },
         "spec": {
            "image": "foo-frontend:1.2.3",
            "port": 3000
         }
      }
   ],
   "kind": "List"
}