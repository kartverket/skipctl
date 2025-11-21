# ArgoKit v2 API Reference

## Jsonnet ArgoKit API

### `argokit.appAndObjects.application.new()`
Oppretter en Skiperator‑applikasjon ved å bruke `appAndObjects`‑konvensjonen (dette er standard).

|navn|type|obligatorisk|standardverdi|beskrivelse|
|-|-|-|-|-|
|`name`|`string`|`true`| - | navn på applikasjonen|
|`image`|`string`|`true`| - | image som skal kjøres|
|`port`|`number`|`true`| - | port applikasjonen skal kjøre på|

**Eksempel:** [examples/application.jsonnet](https://github.com/kartverket/argokit/blob/main/v2/examples/application.jsonnet)

### `argokit.appAndObjects.application.withObjects()`
Legg til ekstra objekter i `objects`-listen til `appAndObjects`-strukturen.

|navn|type|obligatorisk|standardverdi|beskrivelse|
|-|-|-|-|-|
|`objects`|`array` or `object` |`true`| - | objekter som skal legges til i manifestet|

**Eksempel:** [examples/additionalObjects.jsonnet](https://github.com/kartverket/argokit/blob/main/v2/examples/additionalObjects.jsonnet)

## ArgoKit's Replicas API

**OBS!** Det anbefales ikke å kjøre med færre enn 2 replikaer. Dette antallet er satt for at du skal ha god nok tilgjengelighet for applikasjonen din, med å legge til redundans.

### `argokit.appAndObjects.application.withReplicas()`
Sett replikaer for en applikasjon med autoskalering basert på CPU og minne.

|navn|type|obligatorisk|standardverdi|beskrivelse|
|-|-|-|-|-|
|`initial`|`number`|`true`|-|initialt antall replikaer (2 er anbefalt)|
|`max`|`number`|`false`|-|maksimum antall replikaer (hvis satt, aktiverer autoskalering)|
|`targetCpuUtilization`|`number`|`false`|80|CPU-terskel i prosent før autoskalering|
|`targetMemoryUtilization`|`number`|`false`|-|Minneterskel i prosent før autoskalering|

**Eksempel:** [examples/replicas.jsonnet](https://github.com/kartverket/argokit/blob/main/v2/examples/replicas.jsonnet)

## ArgoKit's Environment API

### `argokit.appAndObjects.application.withEnvironmentVariable()`
Oppretter en miljøvariabel for en applikasjon.

|navn|type|obligatorisk|standardverdi|beskrivelse|
|-|-|-|-|-|
|`name`|`string`|`true`| - |navn på miljøvariabelen|
|`value`|`string`|`true`| - |verdi for miljøvariabelen|

**Eksempel:** [examples/environment.jsonnet](https://github.com/kartverket/argokit/blob/main/v2/examples/environment.jsonnet)

### `argokit.appAndObjects.application.withEnvironmentVariables()`
Oppretter flere miljøvariabler for en applikasjon.

|navn|type|obligatorisk|standardverdi|beskrivelse|
|-|-|-|-|-|
|`envVars`|`object`|`true`| - |objekt med nøkkelverdi par for miljøvariabler|

**Eksempel:** [examples/environment.jsonnet](https://github.com/kartverket/argokit/blob/main/v2/examples/environment.jsonnet)

### `argokit.appAndObjects.application.withEnvironmentVariableFromSecret()`
Oppretter en miljøvariabel fra en hemmelighet.

|navn|type|obligatorisk|standardverdi|beskrivelse|
|-|-|-|-|-|
|`name`|`string`|`true`| - |navn på miljøvariabelen|
|`secretRef`|`string`|`true`| - |navn på hemmelighet|
|`key`|`string`|`false`|`name`|nøkkel i hemmelighet (default er samme som `name`)|

**Eksempel:** [examples/environment.jsonnet](https://github.com/kartverket/argokit/blob/main/v2/examples/environment.jsonnet)

### `argokit.appAndObjects.application.withEnvironmentVariablesFromSecret()`
Oppretter miljøvariabler fra en hemmelighet.

|navn|type|obligatorisk|standardverdi|beskrivelse|
|-|-|-|-|-|
|`secretName`|`string`|`true`| - |navn på hemmelighet|

**Eksempel:** [examples/environment.jsonnet](https://github.com/kartverket/argokit/blob/main/v2/examples/environment.jsonnet)

## ArgoKit's Ingress API

### `argokit.appAndObjects.application.forHostnames()`
Oppretter ingress for en applikasjon.

|navn|type|obligatorisk|standardverdi|beskrivelse|
|-|-|-|-|-|
|`ingress`|`array` or `string` or `object`|`true`| - |kan være en string med `hostname`, en array av `hostnames`/objekter, eller et objekt med {`hostname`, `customCert`}|

**Eksempel:** [examples/ingress.jsonnet](https://github.com/kartverket/argokit/blob/main/v2/examples/ingress.jsonnet)

## ArgoKit's accessPolicies API

Du kan definere hvilke eksterne tjenester (verter/IP‑er) og interne SKIP‑applikasjoner appen din kan kommunisere med.

### `argokit.appAndObjects.application.withOutboundPostgres()`
Tillat utgående trafikk til en Postgres‑instans.

|navn|type|obligatorisk|standardverdi|beskrivelse|
|-|-|-|-|-|
|`host`|`string`|`true`| - |vertsnavn til Postgres-instansen|
|`ip`|`string`|`true`| - |IP-adresse til Postgres-instansen|

**Eksempel:** [examples/accessPolicies.jsonnet](https://github.com/kartverket/argokit/blob/main/v2/examples/accessPolicies.jsonnet)

### `argokit.appAndObjects.application.withOutboundOracle()`
Tillat utgående trafikk til en Oracle‑database.

|navn|type|obligatorisk|standardverdi|beskrivelse|
|-|-|-|-|-|
|`host`|`string`|`true`| - |vertsnavn til Oracle-databasen|
|`ip`|`string`|`true`| - |IP-adresse til Oracle-databasen|

**Eksempel:** [examples/accessPolicies.jsonnet](https://github.com/kartverket/argokit/blob/main/v2/examples/accessPolicies.jsonnet)

### `argokit.appAndObjects.application.withOutboundSsh()`
Tillat utgående SSH.

|navn|type|obligatorisk|standardverdi|beskrivelse|
|-|-|-|-|-|
|`host`|`string`|`true`| - |vertsnavn til SSH-serveren|
|`ip`|`string`|`true`| - |IP-adresse til SSH-serveren|

**Eksempel:** [examples/accessPolicies.jsonnet](https://github.com/kartverket/argokit/blob/main/v2/examples/accessPolicies.jsonnet)

### `argokit.appAndObjects.application.withOutboundLdaps()`
Tillat utgående sikker LDAP‑port.

|navn|type|obligatorisk|standardverdi|beskrivelse|
|-|-|-|-|-|
|`host`|`string`|`true`| - |vertsnavn til LDAP-serveren|
|`ip`|`string`|`true`| - |IP-adresse til LDAP-serveren|

**Eksempel:** [examples/accessPolicies.jsonnet](https://github.com/kartverket/argokit/blob/main/v2/examples/accessPolicies.jsonnet)

### `argokit.appAndObjects.application.withOutboundHttp()`
Tillat utgående HTTPS/HTTP til en vert.

|navn|type|obligatorisk|standardverdi|beskrivelse|
|-|-|-|-|-|
|`host`|`string`|`true`| - |vertsnavn til serveren|
|`portname`|`string`|`false`|-|navn på porten|
|`port`|`number`|`false`|443|portnummer|
|`protocol`|`string`|`false`|-|protokoll (HTTP/HTTPS)|

**Eksempel:** [examples/accessPolicies.jsonnet](https://github.com/kartverket/argokit/blob/main/v2/examples/accessPolicies.jsonnet)

### `argokit.appAndObjects.application.withOutboundSkipApp()`
Tillat utgående trafikk til en annen SKIP‑applikasjon (utgående regel).

|navn|type|obligatorisk|standardverdi|beskrivelse|
|-|-|-|-|-|
|appname|`string`|`true`| - |navn på SKIP-applikasjonen|
|namespace|`string`|`false`|-|`namespace` til applikasjonen|

**Eksempel:** [examples/accessPolicies.jsonnet](https://github.com/kartverket/argokit/blob/main/v2/examples/accessPolicies.jsonnet)

### `argokit.appAndObjects.application.withInboundSkipApp()`
Tillat en annen SKIP‑applikasjon å nå denne (inngående regel).

|navn|type|obligatorisk|standardverdi|beskrivelse|
|-|-|-|-|-|
|`appname`|`string`|`true`| - |navn på SKIP-applikasjonen|
|`namespace`|`string`|`false`|-|`namespace` til applikasjonen|

**Eksempel:** [examples/accessPolicies.jsonnet](https://github.com/kartverket/argokit/blob/main/v2/examples/accessPolicies.jsonnet)

## ArgoKit's Probe API
Konfigurer helseprober for applikasjoner.

### `argokit.appAndObjects.application.probe()`
Bygger et probe‑objekt (sti, port, terskler)

|navn|type|obligatorisk|standardverdi|beskrivelse|
|-|-|-|-|-|
|`path`|`string`|`true`| - |sti til probe-endepunktet|
|`port`|`number`|`true`| - |port for probe|
|`failureThreshold`|`number`|`false`|3|antall feil før probe feiler|
|`timeout`|`number`|`false`|1|timeout i sekunder|
|`initialDelay`|`number`|`false`|0|forsinkelse før første probe|

**Eksempel:** [examples/probes.jsonnet](https://github.com/kartverket/argokit/blob/main/v2/examples/probes.jsonnet)

### `argokit.appAndObjects.application.withReadiness()`
Legger til en readiness‑probe (styrer når trafikk sendes til poden).

|navn|type|obligatorisk|standardverdi|beskrivelse|
|-|-|-|-|-|
|`probe`|`object`|`true`| - |probe-objekt opprettet med `probe()`|

**Eksempel:** [examples/probes.jsonnet](https://github.com/kartverket/argokit/blob/main/v2/examples/probes.jsonnet)

### `argokit.appAndObjects.application.withLiveness()`
Legger til en liveness‑probe (restarter container ved feil).

|navn|type|obligatorisk|standardverdi|beskrivelse|
|-|-|-|-|-|
|`probe`|`object`|`true`| - |probe-objekt opprettet med `probe()`|

**Eksempel:** [examples/probes.jsonnet](https://github.com/kartverket/argokit/blob/main/v2/examples/probes.jsonnet)

### `argokit.appAndObjects.application.withStartup()`
Legger til en startup‑probe (blokkerer andre prober til den lykkes).

|navn|type|obligatorisk|standardverdi|beskrivelse|
|-|-|-|-|-|
|`probe`|`object`|`true`| - |probe-objekt opprettet med `probe()`|

**Eksempel:** [examples/probes.jsonnet](https://github.com/kartverket/argokit/blob/main/v2/examples/probes.jsonnet)

## ArgoKit's routing API

Konfigurer ruting for applikasjoner på SKIP.

### `argokit.routing.new()`
Bygger et rute‑objekt.

|navn|type|obligatorisk|standardverdi|beskrivelse|
|-|-|-|-|-|
|`name`|`string`|`true`| - |navn på rute-objektet|
|`hostname`|`string`|`true`| - |vertsnavn for ruten|
|`redirectToHTTPS`|`boolean`|`false`|true|om trafikk skal omdirigeres til HTTPS|

**Eksempel:** [examples/routing.jsonnet](https://github.com/kartverket/argokit/blob/main/v2/examples/routing.jsonnet)

### `argokit.routing.withRoute()`
Legg til rute i `routing`‑objektet.

|navn|type|obligatorisk|standardverdi|beskrivelse|
|-|-|-|-|-|
|`pathPrefix`|`string`|`true`| - |prefix for ruten|
|`targetApp`|`string`|`true`| - |målapplikasjon|
|`rewriteUri`|`boolean`|`true`| - |om URI skal omskrives|
|`port`|`number`|`false`|null|port for målapplikasjonen|

**Eksempel:** [examples/routing.jsonnet](https://github.com/kartverket/argokit/blob/main/v2/examples/routing.jsonnet)

## ArgoKit's Rolebinding API

Konfigurer `rolebinding`‑ressurser for applikasjoner på SKIP. Opprett ressursen med funksjonen `new()`, og legg deretter til enten brukere eller en gruppe som `subject`.

### `argokit.k8s.rolebinding.new()`
Opprett en ny `rolebinding`‑ressurs.

**OBS:** Denne funksjonen må utvides med enten `withUsers()` eller `withNamespaceAdminGroup()` for å være komplett.

|navn|type|obligatorisk|standardverdi|beskrivelse|
|-|-|-|-|-|
|-|-|-|-|ingen parametere|

**Eksempel:** [examples/rolebinding.jsonnet](https://github.com/kartverket/argokit/blob/main/v2/examples/rolebinding.jsonnet)

### `argokit.k8s.rolebinding.withUsers()`
Legg til en liste over brukere som `subjects`.

|navn|type|obligatorisk|standardverdi|beskrivelse|
|-|-|-|-|-|
|`users`|`array`|`true`| - |liste over brukernavn|

**Eksempel:** [examples/rolebinding.jsonnet](https://github.com/kartverket/argokit/blob/main/v2/examples/rolebinding.jsonnet)

### `argokit.k8s.rolebinding.withNamespaceAdminGroup()`
Legg til en `namespace‑admin-group` som `subject`.

|navn|type|obligatorisk|standardverdi|beskrivelse|
|-|-|-|-|-|
|`groupName`|`string`|`true`| - |navn på gruppen|

**Eksempel:** [examples/rolebinding.jsonnet](https://github.com/kartverket/argokit/blob/main/v2/examples/rolebinding.jsonnet)

## ArgoKit's ExternalSecret API

Konfigurer `ExternalSecrets` og `SecretStore`.

### `argokit.externalSecrets.secret.new()`
Opprett en ny ekstern secret.

|navn|type|obligatorisk|standardverdi|beskrivelse|
|-|-|-|-|-|
|`name`|`string`|`true`| - |navn på hemmelighet|
|`secrets`|`array`|`false`|[]|array av secret objekter med {`toKey`, `fromSecret`}|
|`allKeysFrom`|`array`|`false`|[]|array av secret objekter med {`fromSecret`} for å hente alle `keys`|
|`secretStoreRef`|`string`|`false`|'gsm'|navn på `store`|

**OBS:** Enten `secrets` eller `allKeysFrom` må inneholde minst ett element.

**Eksempel:** [examples/externalSecrets.jsonnet](https://github.com/kartverket/argokit/blob/main/v2/examples/externalSecrets.jsonnet)

### `argokit.externalSecrets.store.new()`
Opprett en ny ekstern `SecretStore`.

|navn|type|obligatorisk|standardverdi|beskrivelse|
|-|-|-|-|-|
|`name`|`string`|`false`|'gsm'|navn på `store`|
|`gcpProject`|`string`|`true`| - |GCP prosjekt ID|

**Eksempel:** [examples/externalSecrets.jsonnet](https://github.com/kartverket/argokit/blob/main/v2/examples/externalSecrets.jsonnet)

## ArgoKit's ConfigMap API

Konfigurer `ConfigMap`‑ressurser for applikasjoner på SKIP. Alle metoder har parameteren `addHashToName` for å opprette `ConfigMap` med et unikt navn (hashet suffiks).

### `argokit.k8s.configMap.new()`
Opprett en ny `ConfigMap`.

|navn|type|obligatorisk|standardverdi|beskrivelse|
|-|-|-|-|-|
|`name`|`string`|`true`| - |navn på `ConfigMap`|
|`data`|`object`|`true`| - |data i `ConfigMap`|
|`addHashToName`|`boolean`|`false`|false|om hash skal legges til navnet|

**Eksempel:** [examples/configMap.jsonnet](https://github.com/kartverket/argokit/blob/main/v2/examples/configMap.jsonnet)

### `argokit.appAndObjects.application.withConfigMapAsEnv()`
Opprett en ny `ConfigMap` og legg innholdet som `env` i applikasjonen. Hver nøkkel i `ConfigMap` blir en egen miljøvariabel med en tilsvarende verdi i applikasjonen.  

|navn|type|obligatorisk|standardverdi|beskrivelse|
|-|-|-|-|-|
|`name`|`string`|`true`| - |navn på `ConfigMap`|
|`data`|`object`|`true`| - |data i `ConfigMap`|
|`addHashToName`|`boolean`|`false`|false|om hash skal legges til navnet|

**Eksempel:** [examples/withConfigMap.jsonnet](https://github.com/kartverket/argokit/blob/main/v2/examples/withConfigMap.jsonnet)

### `argokit.appAndObjects.application.withConfigMapAsMount()`
Opprett en ny `ConfigMap` og monter filer i applikasjonens filsystem. Hver nøkkel i `ConfigMap` blir en egen fil med verdien som filinnhold.

|navn|type|obligatorisk|standardverdi|beskrivelse|
|-|-|-|-|-|
|`name`|`string`|`true`| - |navn på `ConfigMap`|
|`mountPath`|`string`|`true`| - |`mountPath` i container|
|`data`|`object`|`true`| - |data i `ConfigMap`|
|`addHashToName`|`boolean`|`false`|false|om hash skal legges til navnet|

**Eksempel:** [examples/withConfigMap.jsonnet](https://github.com/kartverket/argokit/blob/main/v2/examples/withConfigMap.jsonnet)

## ArgoKit's ExternalSecrets API


### `argokit.appAndObjects.application.withEnvironmentVariablesFromExternalSecret()`
Opprett en `ExternalSecret` og legg til miljøvariabler fra den i applikasjonen.

|navn|type|obligatorisk|standardverdi|beskrivelse|
|-|-|-|-|-|
|`name`|`string`|`true`| - |navn på `ExternalSecret`|
|`secrets`|`array`|`false`|[]|array av secret objekter med {`toKey`, `fromSecret`}|
|`allKeysFrom`|`array`|`false`|[]|array av secret objekter med {`fromSecret`} for å hente alle `keys`|
|`secretStoreRef`|`string`|`false`|'gsm'|navn på `store`|

**OBS:** Enten `secrets` eller `allKeysFrom` må inneholde minst ett element.

**Eksempel:** [examples/externalSecrets.jsonnet](https://github.com/kartverket/argokit/blob/main/v2/examples/externalSecrets.jsonnet)

## ArgoKit's AzureAD API

### `argokit.azureAdApplication.new()`
Opprett en frittstående `AzureADApplication`-ressurs.

|navn|type|obligatorisk|standardverdi|beskrivelse|
|-|-|-|-|-|
|`name`|`string`|`true`| - |navn på `AzureAdApplication`-ressursen|
|`namespace`|`string`|`false`|-|`namespace` for ressursen|
|`groups`|`array`|`false`|[]|Azure AD grupper for `claims`|
|`secretPrefix`|`string`|`false`|'azuread'|prefix for `secret`-navnet|
|`allowAllUsers`|`boolean`|`false`|false|om alle brukere skal ha tilgang|
|`logoutUrl`|`string`|`false`|-|logout URL|
|`replyUrls`|`array`|`false`|[]|liste over reply URLs|
|`preAuthorizedApplications`|`array`|`false`|[]|liste over forhåndsautoriserte applikasjoner|

**Eksempel:** [examples/newAzureAdApplication.jsonnet](https://github.com/kartverket/argokit/blob/main/v2/examples/newAzureAdApplication.jsonnet)



### `argokit.appAndObjects.application.withAzureAdApplication()`
Legg til en `AzureADApplication`-ressurs og konfigurerer applikasjonen.

|navn|type|obligatorisk|standardverdi|beskrivelse|
|-|-|-|-|-|
|`name`|`string`|`true`| - |navn på `AzureAdApplication`-ressursen|
|`namespace`|`string`|`false`|-|`namespace` for ressursen|
|`groups`|`array`|`false`|[]|Azure AD grupper for `claims`|
|`secretPrefix`|`string`|`false`|'azuread'|prefix for `secret` navnet|
|`allowAllUsers`|`boolean`|`false`|false|om alle brukere skal ha tilgang|
|`logoutUrl`|`string`|`false`|-|logout URL|
|`replyUrls`|`array`|`false`|[]|liste over reply URLs|
|`preAuthorizedApplications`|`array`|`false`|[]|liste over forhåndsautoriserte applikasjoner|

**Eksempel:** [examples/withAzureAdApplication.jsonnet](https://github.com/kartverket/argokit/blob/main/v2/examples/withAzureAdApplication.jsonnet)
