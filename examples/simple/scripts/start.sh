service mysql start
mysql -u root -e "CREATE DATABASE IF NOT EXISTS gin_job;"
mysql -u root -e "USE mysql;ALTER USER 'root'@'localhost' IDENTIFIED WITH mysql_native_password BY '$MYSQL_ROOT_PASSWORD';FLUSH PRIVILEGES;"
/app/gin-job