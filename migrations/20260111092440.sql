-- Create "gateway_requests" table
CREATE TABLE "public"."gateway_requests" (
  "id" bigserial NOT NULL,
  "created_at" timestamptz NULL,
  "updated_at" timestamptz NULL,
  "uuid" uuid NOT NULL,
  "request_data" text NULL,
  PRIMARY KEY ("id")
);
-- Create index "idx_gateway_requests_uuid" to table: "gateway_requests"
CREATE UNIQUE INDEX "idx_gateway_requests_uuid" ON "public"."gateway_requests" ("uuid");
