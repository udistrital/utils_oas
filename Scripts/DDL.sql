--Script Creacion DB Athena

CREATE DATABASE IF NOT EXISTS logswso2
  COMMENT 'DB logs wso2'
  WITH DBPROPERTIES ('creator'='Auditoria', 'Dept.'='CORE');

--Script Creacion Tabla de formato de archivos Json almacenados en S3

CREATE EXTERNAL TABLE IF NOT EXISTS logswso2.registro_logs_v2 (
  `tid` string,
  `date` string,
  `level` string,
  `package` string,
  `details` string,
  `ip` string,
  `host` string 
)
ROW FORMAT SERDE 'org.apache.hadoop.hive.serde2.lazy.LazySimpleSerDe'
WITH SERDEPROPERTIES (
  'serialization.format' = '|',
  'field.delim' = '|',
  'collection.delim' = '@',
  'mapkey.delim' = '%'
) LOCATION 's3://logs-wso2-oas/'
TBLPROPERTIES ('has_encrypted_data'='false');

--Descripcion de campos 

-- ddl-end --
COMMENT ON COLUMN logswso2.registro_logs_v2.tid IS 'Identificador pendiente de definicion';
-- ddl-end --
COMMENT ON COLUMN logswso2.registro_logs_v2.date IS 'Fecha en que se genero el registro de auditoria';
-- ddl-end --
COMMENT ON COLUMN logswso2.registro_logs_v2.level IS 'Identificador del tipo de registro generado, posibles valores ERROR, INFO';
-- ddl-end --
COMMENT ON COLUMN logswso2.registro_logs_v2.package IS 'Los tipos de valores que agrupa este campo y su utilidad actualmente son desconocidos';
-- ddl-end --
COMMENT ON COLUMN logswso2.registro_logs_v2.details IS 'Registro de la accion que se pretende ejecutar ya sea una descripcion del error o el payload generado desde los API''s';
-- ddl-end --
COMMENT ON COLUMN logswso2.registro_logs_v2.ip IS 'Direccion IP del ordenador en que se hizo la peticion';
-- ddl-end --
COMMENT ON COLUMN logswso2.registro_logs_v2.host IS 'endpoint de la solicitud';
-- ddl-end --