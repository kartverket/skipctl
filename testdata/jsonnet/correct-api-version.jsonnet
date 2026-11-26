{
  // apiVersion: 'skiperator.kartverket.no/v1beta1',  // only v1alpha1 is supported
  apiVersion: 'skiperator.kartverket.no/v1alpha1',
  kind: 'Application',
  metadata: {
    name: 'test-app',
  },
  spec: {
    image: 'some-image',
    port: 3000,
  },
}