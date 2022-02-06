--Script Creacion DB Athena

CREATE DATABASE IF NOT EXISTS logsapis
  COMMENT 'DB logs apis'
  WITH DBPROPERTIES ('creator'='Auditoria', 'Dept.'='CORE');

--Script Creacion Tabla formato de archivos almacenados en S3

CREATE EXTERNAL TABLE `logs_apis`(
  `c_info` string COMMENT 'Informacion consola API auditoria generada por BeeGo',
  `app_name` string COMMENT 'Nombre del API al que se le hace la peticion',
  `host` string COMMENT 'Host del API',
  `end_point` string COMMENT 'Endpoint al que se le realiza la peticion',
  `method` string COMMENT 'Metodo REST de la peticion',
  `date` string COMMENT 'Fecha y hora de la operacion',
  `ip_user` string COMMENT 'IP del usuario',
  `user_agent` string COMMENT 'Tipo de aplicacion, sistema operativo, proveedor del software o la version del software de la peticion del agente de usuario',
  `user` string COMMENT 'Nombre de usuario en WSO2 que realiza la peticion',
  `data_response` string COMMENT 'Payload del servicio')
ROW FORMAT SERDE
  'org.apache.hadoop.hive.serde2.RegexSerDe'
WITH SERDEPROPERTIES (
  'input.regex'='(.+?)@&(.+?)@&(.+?)@&(.+?)@&(.+?)@&(.+?)@&(.+?)@&(.+?)@&(.+?)@&([^\n]*)')
STORED AS INPUTFORMAT
  'org.apache.hadoop.mapred.TextInputFormat'
OUTPUTFORMAT
  'org.apache.hadoop.hive.ql.io.HiveIgnoreKeyTextOutputFormat'
LOCATION
  's3://auditoria-logs-apis/'
TBLPROPERTIES (
  'has_encrypted_data'='false',
  'transient_lastDdlTime'='1561650804')
