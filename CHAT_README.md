# Skipctl Chat - Mai

Chat med Mai for å jobbe med Kubernetes manifester og ArgoKit.

## Hva er dette?

Mai er en innebygd AI-assistent i Skipctl som hjelper deg med:
- 📝 Validere manifester
- 🔍 Vise hva som vil deployes (render)
- 📊 Sammenligne endringer (diff)
- ✨ Formatere filer
- 🗂️ Utforske prosjektstrukturen
- 💡 Forklare ArgoKit-konsepter

## Oppsett (2 minutt)

### 1. Få API-nøkkel

Gå til https://console.anthropic.com/ og lag en API-nøkkel.

### 2. Sett miljøvariabel

```bash
export ANTHROPIC_API_KEY=sk-ant-...
```

Legg den i `~/.zshrc` eller `~/.bashrc` for permanent oppsett:
```bash
echo 'export ANTHROPIC_API_KEY=sk-ant-...' >> ~/.zshrc
source ~/.zshrc
```

### 3. Ferdig!

```bash
skipctl ai "what manifests do I have?"
```

## Bruk

### Enkel modus (single-shot)

Still ett spørsmål og få svar:

```bash
# List manifester
skipctl chat "what manifests do I have?"

# Valider en fil
skipctl chat "is testdata/yaml/valid.yaml valid?"

# Vis hva som deployes
skipctl chat "show me what testdata/yaml/valid.yaml will deploy"

# Sammenlign endringer
skipctl chat "show me the diff for testdata/yaml/valid.yaml against HEAD"
```

### Interaktiv modus (chat)

Start en samtale:

```bash
skipctl chat
```

Eksempel samtale:
```
💬 Mai - Skipctl Chat
Ask me anything about your manifests. Type 'exit' to quit.

You: what manifests do I have in testdata?
Mai: I found 6 manifest files in testdata:
- testdata/yaml/valid.yaml
- testdata/yaml/invalid.yaml
...

You: validate testdata/yaml/valid.yaml
Mai: ✅ Manifest testdata/yaml/valid.yaml is valid

You: what will it deploy?
Mai: This manifest will deploy an Application called "valid-manifest" 
in the devex namespace. It will:
- Use the kartverket/example image
- Expose port 8080
- Run between 1 and 5 replicas

You: exit
Goodbye! 👋
```

## Eksempler

### Validering av alle filer

```bash
skipctl chat "validate all manifests in testdata/yaml"
```

### Forstå en manifest

```bash
skipctl chat "explain what testdata/yaml/valid.yaml does"
```

### Sammenligning med main branch

```bash
skipctl chat "show me what changed in testdata/yaml/valid.yaml compared to HEAD"
```

### Hjelp med ArgoKit

```bash
skipctl chat "how do I add a Redis cache to my ArgoKit application?"
```

## Hvordan fungerer det?

1. **Du stiller et spørsmål** på vanlig norsk eller engelsk
2. **Mai analyserer spørsmålet** og bestemmer hvilke tools hun trenger
3. **Mai kaller Skipctl-verktøy**:
   - `list_manifests` - finner filer
   - `render_manifest` - viser YAML-output
   - `diff_manifest` - sammenligner versjoner
   - `validate_manifest` - sjekker syntaks
   - `format_manifest` - rydder filer
4. **Mai formulerer svaret** basert på resultatene

## Kostnader

Claude API koster penger, men er billig for vanlig bruk:

- Input: ~$3 per million tokens
- Output: ~$15 per million tokens

**Typisk bruk:**
- "what manifests do I have?" ≈ $0.001
- "validate all my manifests" ≈ $0.005
- En hel samtale (10 meldinger) ≈ $0.02

**Tips for å spare penger:**
- Bruk `--model claude-3-haiku-20240307` for billigere modell
- Still spesifikke spørsmål
- Unngå å sende samme spørsmål flere ganger

## Feilsøking

### "No API key found"

Sett miljøvariabelen:
```bash
export ANTHROPIC_API_KEY=sk-ant-...
```

Sjekk at den er satt:
```bash
echo $ANTHROPIC_API_KEY
```

### "API error (status 401)"

API-nøkkelen er ugyldig. Lag en ny på https://console.anthropic.com/

### "API error (status 429)"

Du har brukt opp rate limit eller kreditter. Sjekk kontoen din.

### Chat gir feil filsti

Bruk absolutt eller relativ sti fra der du står:
```bash
# Fra /Users/even/git/skipctl/
skipctl chat "validate testdata/yaml/valid.yaml"

# Eller absolutt sti
skipctl chat "validate /Users/even/git/skipctl/testdata/yaml/valid.yaml"
```

## Sammenlignet med MCP

| Funksjon | `skipctl chat` | MCP + Claude Desktop |
|----------|--------------|----------------------|
| Oppsett | ✅ Enkelt (1 env var) | ⚠️ Krever Claude Desktop config |
| Bruk | ✅ Direkte i terminal | ⚠️ Må bruke Claude Desktop app |
| Samtalehistorikk | ⚠️ Kun i interaktiv modus | ✅ Full historikk |
| API-kostnad | 💰 Du betaler | 💰 Du betaler |
| Offline | ❌ Nei | ❌ Nei |

**Anbefaling:**
- Bruk `skipctl chat` for raske spørsmål i terminalen
- Bruk MCP + Claude Desktop for lengre samtaler og research

## Eksempel workflow

### Før deployment

```bash
# Start interaktiv modus
skipctl chat

# Sjekk hvilke filer som endret seg
> what files changed in manifests/ compared to main?

# Valider alle
> validate all those files

# Vis diff for viktige filer
> show me the diff for manifests/prod/app.yaml against main

# Forklarer endringene
> explain what will happen if I deploy these changes

> exit
```

### Lære ArgoKit

```bash
skipctl chat "how do I configure resource limits in ArgoKit?"
skipctl chat "show me an example of an ArgoKit application with ingress"
skipctl chat "what's the difference between Application and SKIPJob?"
```

## Tips

1. **Vær spesifikk**: "validate testdata/yaml/valid.yaml" > "validate my file"
2. **Bruk interaktiv modus** for relaterte spørsmål (deler kontekst)
3. **Still oppfølgingsspørsmål**: "explain that more" "what about..."
4. **Be om eksempler**: "show me an example"
5. **Kombiner operasjoner**: "validate and show me the diff"

## Sikkerhet

⚠️ **OBS:**
- Chat sender filinnhold til Claude API (Anthropic)
- Ikke bruk på filer med hemmeligheter/passord
- API-kallet går over HTTPS (kryptert)
- Anthropic lagrer ikke data i over 30 dager

Les mer: https://www.anthropic.com/legal/privacy

### Sikkerhetsfunksjoner

Skipctl chat har innebygd sikkerhet for å beskytte systemet ditt:

#### Standard sikkerhet (default)
```bash
skipctl chat "validate my manifests"
```

- ✅ Kun tilgang til filer i nåværende mappe (working directory)
- ✅ Maks filstørrelse: 10MB
- ✅ Skrivoperasjoner tillatt (format)
- ✅ Rate limiting: 20 API-kall per minutt

#### Streng sikkerhet (strict)
```bash
skipctl chat --security=strict "validate my manifests"
```

- 🔒 Kun tilgang til filer i nåværende mappe
- 🔒 Maks filstørrelse: 1MB
- 🔒 **Ingen skrivoperasjoner** (read-only)
- 🔒 Rate limiting: 20 API-kall per minutt

Bruk for produksjonsfiler eller når du er usikker.

#### Avslappet sikkerhet (relaxed)
```bash
skipctl chat --security=relaxed "validate ~/projects/other/manifest.yaml"
```

- ⚠️ Tilgang til filer utenfor working directory
- ⚠️ Maks filstørrelse: 50MB
- ⚠️ Skrivoperasjoner tillatt
- ⚠️ Rate limiting: 20 API-kall per minutt

Bruk kun når du trenger tilgang til filer andre steder.

#### Egendefinert sikkerhet
```bash
# Deaktiver skriving (read-only)
skipctl chat --read-only "format my manifests"

# Tillat større filer (20MB)
skipctl chat --max-file-size=20971520 "validate large-manifest.yaml"

# Tillat tilgang utenfor working directory
skipctl chat --allow-all-paths "validate ~/other-project/manifest.yaml"

# Kombinasjoner
skipctl chat --security=strict --allow-all-paths "list manifests in /tmp"
```

### Beskyttelse mot farlige operasjoner

Chat-kommandoen beskytter mot:

1. **Path traversal**: Kan ikke lese `/etc/passwd`, `../../secrets.txt`
2. **Systemfiler**: Blokkerer tilgang til `/etc/`, `/var/`, `/System/`
3. **For store filer**: Forhindrer å laste inn gigantiske filer
4. **Uautorisert skriving**: Kan deaktiveres helt
5. **API misbruk**: Rate limiting forhindrer for mange kall

### Anbefalinger

**For daglig bruk:**
```bash
skipctl chat "validate my manifests"  # Standard sikkerhet
```

**For produksjonsfiler:**
```bash
skipctl chat --security=strict "validate prod-manifests/"
```

**Når du vil formatere filer:**
```bash
# Default tillater skriving
skipctl chat "format testdata/yaml/formatme.yml"
```

**Når du IKKE vil at Mai skal endre noe:**
```bash
skipctl chat --read-only "help me understand this manifest"
```
