CREATE TABLE IF NOT EXISTS names (
  id    SERIAL,
  country_letter varchar(4) UNIQUE NOT NULL,  
  country_name varchar(90) NOT NULL,  
  PRIMARY KEY (id)
);

CREATE TABLE IF NOT EXISTS phone (
  id    SERIAL,
  country_letter varchar(4) UNIQUE NOT NULL REFERENCES names(country_letter),  
  country_code varchar(50) NOT NULL,  
  PRIMARY KEY (id)  
);
