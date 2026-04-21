ALTER TABLE devices ADD COLUMN display_name TEXT;
ALTER TABLE devices ADD COLUMN device_type TEXT;
ALTER TABLE devices ADD COLUMN classification_confidence TEXT;
ALTER TABLE devices ADD COLUMN classification_reasons TEXT;

ALTER TABLE device_observations ADD COLUMN display_name TEXT;
ALTER TABLE device_observations ADD COLUMN device_type TEXT;
ALTER TABLE device_observations ADD COLUMN classification_confidence TEXT;
ALTER TABLE device_observations ADD COLUMN classification_reasons TEXT;
