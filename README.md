# ⚽ Golazo API

API backend de una plataforma de apuestas deportivas, inspirada en las casas de apuestas peruanas. Está escrita en **Go** con **PostgreSQL** y se construye por etapas: empieza como un monolito y evolucionará hacia un sistema distribuido.

> Proyecto de aprendizaje y portafolio. No maneja dinero real.

## Qué hace hoy

- Registro de usuarios con validaciones (DNI, email, celular peruano, mayoría de edad).
- Login con JWT.
- Billetera en soles (PEN) creada automáticamente al registrarse.
- Depósitos protegidos contra cobros duplicados (idempotencia) y concurrencia.
- Historial de movimientos (ledger) que siempre cuadra con el saldo.

## Qué tendrá

- Consulta de saldo e historial de movimientos.
- Retiros.
- Eventos deportivos, cuotas y apuestas (simples y combinadas).
- Liquidación automática de apuestas con colas (AWS SQS).
- Cuotas en tiempo real.
- Despliegue en AWS con CI/CD.

---

## Requisitos

| Herramienta | Para qué |
|---|---|
| [Go 1.26+](https://go.dev/dl/) | Compilar y ejecutar la API |
| [Docker Desktop](https://www.docker.com/products/docker-desktop/) | Levantar PostgreSQL |
| [golang-migrate](https://github.com/golang-migrate/migrate) | Crear las tablas |
| [Task](https://taskfile.dev/) | Atajos de comandos (`task run`, `task test`...) |

Instala `golang-migrate` y `Task` con Go:

```bash
go install -tags "postgres" github.com/golang-migrate/migrate/v4/cmd/migrate@latest
go install github.com/go-task/task/v3/cmd/task@latest
```

Comprueba que quedaron instalados:

```bash
migrate -version
task --version
```

> Si no se reconocen los comandos, agrega la carpeta `bin` de Go a tu `PATH` (la ves con `go env GOPATH`) y abre una terminal nueva.

---

## Cómo arrancar el proyecto

### 1. Clonar el repositorio

```bash
git clone https://github.com/kingperseocs/golazo-api.git
cd golazo-api
```

### 2. Crear el archivo `.env`

Copia el archivo de ejemplo:

```bash
# Linux / macOS
cp .env.example .env

# Windows (PowerShell)
Copy-Item .env.example .env
```

Abre `.env` y completa los valores. Para desarrollo local, el usuario y la contraseña de Postgres son los del `docker-compose.yml` (`postgres` / `admin`):

```env
APP_ENV=development
HTTP_PORT=8080
DATABASE_URL=postgres://postgres:admin@localhost:5433/golazo?sslmode=disable
TEST_DATABASE_URL=postgres://postgres:admin@localhost:5433/golazo_test?sslmode=disable
JWT_SECRET=pon-aqui-una-clave-larga-y-aleatoria-de-32-caracteres-o-mas
```

> `JWT_SECRET` debe tener **al menos 32 caracteres**; si no, la API no arranca.

### 3. Encender PostgreSQL con Docker

Con Docker Desktop abierto:

```bash
task db:up
```

Espera unos segundos y verifica que esté listo:

```bash
docker compose ps
```

La columna `STATUS` debe decir **`healthy`**.

> Postgres queda expuesto en el puerto **5433** de tu PC, para no chocar con otro Postgres instalado localmente.

### 4. Crear las tablas

```bash
task migrate:up
```

### 5. Levantar la API

```bash
task run
```

Verás en la terminal:

```json
{"level":"INFO","msg":"servidor HTTP escuchando","addr":"[::]:8080","env":"development"}
```

Comprueba que responde:

```bash
curl http://localhost:8080/health
```

Para detener la API, presiona `Ctrl + C`.

---

## Tests

**Tests unitarios** (no necesitan Docker):

```bash
task test
```

**Tests de integración** (contra PostgreSQL real). La primera vez, crea la base de datos de tests:

```bash
docker exec golazo-postgres psql -U postgres -c "CREATE DATABASE golazo_test;"
```

Y después, cada vez que quieras ejecutarlos:

```bash
task test:integration
```

---

## Comandos disponibles

| Comando | Qué hace |
|---|---|
| `task` | Lista todos los comandos |
| `task db:up` | Enciende PostgreSQL |
| `task db:stop` | Apaga PostgreSQL (los datos se conservan) |
| `task migrate:up` | Aplica las migraciones pendientes |
| `task migrate:down` | Revierte la última migración |
| `task migrate:create -- nombre` | Crea una migración nueva |
| `task run` | Levanta la API |
| `task test` | Tests unitarios |
| `task test:integration` | Tests de integración |

---

## Endpoints disponibles

| Método | Ruta | Requiere token |
|---|---|---|
| `GET` | `/health` | No |
| `POST` | `/users` | No |
| `POST` | `/auth/login` | No |
| `POST` | `/wallets/{id}/deposits` | Sí (`Authorization: Bearer <token>`) |
