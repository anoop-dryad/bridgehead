-- Create "sensor_requests" table
CREATE TABLE "public"."sensor_requests" (
  "id" bigserial NOT NULL,
  "created_at" timestamptz NULL,
  "updated_at" timestamptz NULL,
  "uuid" uuid NOT NULL,
  "request_data" text NULL,
  PRIMARY KEY ("id")
);
-- Create index "idx_sensor_requests_uuid" to table: "sensor_requests"
CREATE UNIQUE INDEX "idx_sensor_requests_uuid" ON "public"."sensor_requests" ("uuid");
