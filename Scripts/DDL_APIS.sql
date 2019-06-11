--Script Creacion DB Athena

CREATE DATABASE IF NOT EXISTS logsapis
  COMMENT 'DB logs apis'
  WITH DBPROPERTIES ('creator'='Auditoria', 'Dept.'='CORE');

--Script Creacion Tabla formato de archivos almacenados en S3

CREATE EXTERNAL TABLE IF NOT EXISTS logs.logs_apis_test (
  `app_name` string COMMENT 'Nombre del API al que se le hace la peticion',
  `host` string COMMENT 'Host del API',
  `end_point` string COMMENT 'Endpoint al que se le realiza la peticion',
  `method` string COMMENT 'Metodo REST de la peticion',
  `date` string COMMENT 'Fecha y hora de la operacion',
  `ip_user` string COMMENT 'IP del usuario',
  `user_agent` string COMMENT 'Tipo de aplicacion, sistema operativo, proveedor del software o la version del software de la peticion del agente de usuario',
  `user` string COMMENT 'Nombre de usuario en WSO2 que realiza la peticion',
  `data_response` string COMMENT 'Payload del servicio' 
)
ROW FORMAT SERDE 'org.apache.hadoop.hive.serde2.RegexSerDe'
WITH SERDEPROPERTIES (
  'serialization.format' = '1',
  'input.regex' = '([^\\r\\n]*)@&([^\\r\\n]*)@&([^\\r\\n]*)@&([^\\r\\n]*)@&([^\\r\\n]*)@&([^\\r\\n]*)@&([^\\r\\n]*)@&([^\\r\\n]*)@&([^\\r\\n]*)'
) LOCATION 's3://logs-wso2-oas-2/cleaned_logs/'
TBLPROPERTIES ('has_encrypted_data'='false');