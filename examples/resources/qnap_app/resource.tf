resource "qnap_app" "postgresql-test" {
  name              = "postgresql-test"
  removeanonvolumes = true
  yml               = "version: '3'\nservices:\n  postgres:\n    image: postgres:15.1\n    restart: always\n    ports:\n      - 127.0.0.1:5432:5432\n    volumes:\n      - postgres_db:/var/lib/postgresql/data\n    environment:\n      POSTGRES_USER: postgres_qnap_user\n      POSTGRES_PASSWORD: postgres_qnap_pwd\n\n  phppgadmin:\n    image: qnapsystem/phppgadmin:7.13.0-1\n    restart: on-failure\n    ports:\n      - 7070:80\n    depends_on:\n      - postgres\n    environment:\n      PHP_PG_ADMIN_SERVER_HOST: postgres\n      PHP_PG_ADMIN_SERVER_PORT: 5432\n\nvolumes:\n  postgres_db:\n"
}