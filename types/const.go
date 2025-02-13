package types

const DEFAULT_API_BASE = "https://api.locker.io/locker_secrets"
const ENCRYPTED_DATA = "encrypted_data_block.json"
const REVISION_DATE = "revision_date.json"
const CREDENTIAL = "credential.json"

const ERR_NOT_FOUND = "not_found_error"
const ERR_FUNC = "function_error"
const ERR_FLAG = "flag_error"
const ERR_INPUT = "input_error"
const ERR_INPUT_KEY = "invalid_secret_access_key"
const ERR_DATA = "data_error"
const ERR_SERVER = "server_error"
const ERR_HTTP = "http_error"
const ERR_FILE = "file_error"
const ERR_PATH = "path_error"
const ERR_DB = "database_error"

const SERVER_ERR_MSG_DUP = "hash already exists"

const FETCH_KIND_RUN = "run"
const FETCH_KIND_SEC = "secrets"
const FETCH_KIND_ENV = "environments"
const FETCH_KIND_PROFILE = "profile"
const FETCH_KIND_FORCE = "force"

const REG_ACCESS_KEY_ID = "LOCKER_ACCESS_KEY_ID"
const REG_ACCESS_KEY_SECRET = "LOCKER_ACCESS_KEY_SECRET"

const DB_REVISION_NUMBER = 1

const OPERATION_CREATE = "CREATE"
const OPERATION_UPDATE = "UPDATE"
