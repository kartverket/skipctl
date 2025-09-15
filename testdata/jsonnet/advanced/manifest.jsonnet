local application = import './func.libsonnet';


application(
  env='dev',
  version='v1.2.1',
  clientId='xyz',
  port=3333
)
