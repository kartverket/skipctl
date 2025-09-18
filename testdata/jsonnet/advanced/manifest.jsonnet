local application = import './func.libsonnet';


application(
  env='dev',
  version='v1.2.3',
  clientId='xyz',
  port=3333,
  ingress='hostname',
)
