CREATE TABLE tractor_unit (
                              id BIGSERIAL PRIMARY KEY,
                              license_plate VARCHAR(20) NOT NULL,
                              driver_id BIGINT NOT NULL,
                              user_id BIGINT NOT NULL,
                              chassis VARCHAR(50) NOT NULL,
                              brand VARCHAR(50) NOT NULL,
                              model VARCHAR(50) NOT NULL,
                              manufacture_year INT,
                              engine_power VARCHAR(50),
                              unit_type VARCHAR(50),
                              can_couple BOOLEAN,
                              height FLOAT,
                              status BOOL not null,
                              created_at timestamp not null,
                              updated_at timestamp null,
                              CONSTRAINT fk_driver
                              FOREIGN KEY (driver_id) REFERENCES driver(id),
                              CONSTRAINT fk_user
                              FOREIGN KEY (user_id) REFERENCES users(id)
);

CREATE TABLE trailer (
                         id BIGSERIAL PRIMARY KEY,
                         license_plate VARCHAR(20) NOT NULL,
                         user_id BIGINT NOT NULL,
                         chassis VARCHAR(50) NOT NULL,
                         body_type VARCHAR(50),
                         load_capacity FLOAT,
                         length FLOAT,
                         width FLOAT,
                         height FLOAT,
                         status BOOL not null,
                         created_at timestamp not null,
                         updated_at timestamp null,
                         CONSTRAINT fk_user
                         FOREIGN KEY (user_id) REFERENCES users(id)
);


CREATE TABLE truck (
                          id BIGINT PRIMARY KEY,
                          tractor_unit_id BIGINT NOT NULL,
                          trailer_id  BIGINT,
                          driver_id BIGINT NOT NULL,
                          CONSTRAINT fk_tractor_unit FOREIGN KEY (tractor_unit_id) REFERENCES tractor_unit(id),
                          CONSTRAINT fk_trailer FOREIGN KEY (trailer_id) REFERENCES trailer(id),
                          CONSTRAINT fk_driver FOREIGN KEY (driver_id) REFERENCES driver(id)
);
