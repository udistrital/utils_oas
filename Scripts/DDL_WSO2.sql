--Script Creacion DB Athena

CREATE DATABASE IF NOT EXISTS logswso2
  COMMENT 'DB logs wso2'
  WITH DBPROPERTIES ('creator'='Auditoria', 'Dept.'='CORE');

--Script Creacion Tabla de formato de archivos Json almacenados en S3

CREATE EXTERNAL TABLE `logs_wso2`(
  `tid` string COMMENT 'Identificador generado en la consola WSO2',
  `ip` string COMMENT 'Direccion IP del ordenador en que se hizo la peticion',
  `host` string COMMENT 'Endpoint de la solicitud',
  `date` string COMMENT 'Fecha en que se genero el registro de auditoria',
  `level` string COMMENT 'Identificador del tipo de registro generado, posibles valores ERROR, INFO',
  `package` string COMMENT 'Identificador package',
  `details_opc` string COMMENT 'Parte de la columna Details que se genera en algunos casos de error, en la mayoria de casos estara vacio este campo',
  `details` string COMMENT 'Registro de la accion que se ejecuta ya sea una descripcion del error o el payload generado desde los APIs')
ROW FORMAT SERDE
  'org.apache.hadoop.hive.serde2.RegexSerDe'
WITH SERDEPROPERTIES (
  'input.regex'='([^\\r\\n]*)@&([^\\r\\n]*)@&([^\\r\\n]*)@&([^\\r\\n]*)@&([^\\r\\n]*)@&([^\\r\\n]*)@&([^\\r\\n]*)@&([^\\r\\n]*)')
STORED AS INPUTFORMAT
  'org.apache.hadoop.mapred.TextInputFormat'
OUTPUTFORMAT
  'org.apache.hadoop.hive.ql.io.HiveIgnoreKeyTextOutputFormat'
LOCATION
  's3://auditoria-logs-wso2/'
TBLPROPERTIES (
  'has_encrypted_data'='false',
  'transient_lastDdlTime'='1561651623')
