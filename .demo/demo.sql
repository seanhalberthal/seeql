-- Demo database for seeql screenshots/gifs
-- E-commerce schema: customers, products, orders, order_items, reviews

PRAGMA foreign_keys = ON;

DROP TABLE IF EXISTS reviews;
DROP TABLE IF EXISTS order_items;
DROP TABLE IF EXISTS orders;
DROP TABLE IF EXISTS products;
DROP TABLE IF EXISTS customers;

CREATE TABLE customers (
    id          INTEGER PRIMARY KEY,
    name        TEXT NOT NULL,
    email       TEXT NOT NULL UNIQUE,
    country     TEXT NOT NULL,
    signup_date DATE NOT NULL,
    is_active   INTEGER NOT NULL DEFAULT 1
);

CREATE INDEX idx_customers_country ON customers(country);
CREATE INDEX idx_customers_signup  ON customers(signup_date);

CREATE TABLE products (
    id          INTEGER PRIMARY KEY,
    sku         TEXT NOT NULL UNIQUE,
    name        TEXT NOT NULL,
    category    TEXT NOT NULL,
    price       REAL NOT NULL,
    stock       INTEGER NOT NULL,
    metadata    TEXT,        -- JSON blob
    created_at  DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP
);

CREATE INDEX idx_products_category ON products(category);
CREATE INDEX idx_products_price    ON products(price);

CREATE TABLE orders (
    id          INTEGER PRIMARY KEY,
    customer_id INTEGER NOT NULL,
    status      TEXT NOT NULL CHECK (status IN ('pending','paid','shipped','delivered','cancelled','refunded')),
    total       REAL NOT NULL,
    currency    TEXT NOT NULL DEFAULT 'GBP',
    created_at  DATETIME NOT NULL,
    notes       TEXT,
    FOREIGN KEY (customer_id) REFERENCES customers(id)
);

CREATE INDEX idx_orders_customer ON orders(customer_id);
CREATE INDEX idx_orders_status   ON orders(status);
CREATE INDEX idx_orders_created  ON orders(created_at);

CREATE TABLE order_items (
    id         INTEGER PRIMARY KEY,
    order_id   INTEGER NOT NULL,
    product_id INTEGER NOT NULL,
    quantity   INTEGER NOT NULL,
    unit_price REAL NOT NULL,
    FOREIGN KEY (order_id)   REFERENCES orders(id),
    FOREIGN KEY (product_id) REFERENCES products(id)
);

CREATE INDEX idx_order_items_order   ON order_items(order_id);
CREATE INDEX idx_order_items_product ON order_items(product_id);

CREATE TABLE reviews (
    id          INTEGER PRIMARY KEY,
    product_id  INTEGER NOT NULL,
    customer_id INTEGER,
    rating      INTEGER NOT NULL CHECK (rating BETWEEN 1 AND 5),
    title       TEXT,
    body        TEXT,
    created_at  DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
    FOREIGN KEY (product_id)  REFERENCES products(id),
    FOREIGN KEY (customer_id) REFERENCES customers(id)
);

CREATE INDEX idx_reviews_product  ON reviews(product_id);
CREATE INDEX idx_reviews_rating   ON reviews(rating);

-- ============================================================
-- Customers (40)
-- ============================================================
INSERT INTO customers (id, name, email, country, signup_date, is_active) VALUES
 (1, 'Ava Thompson',    'ava.thompson@example.com',    'UK',     '2023-01-14', 1),
 (2, 'Liam Patel',      'liam.patel@example.com',      'UK',     '2023-02-03', 1),
 (3, 'Sofia Garcia',    'sofia.garcia@example.com',    'ES',     '2023-02-19', 1),
 (4, 'Noah Kim',        'noah.kim@example.com',        'KR',     '2023-03-08', 1),
 (5, 'Isla Murray',     'isla.murray@example.com',     'UK',     '2023-03-22', 1),
 (6, 'Mateo Rossi',     'mateo.rossi@example.com',     'IT',     '2023-04-11', 1),
 (7, 'Zara Ahmed',      'zara.ahmed@example.com',      'UK',     '2023-04-29', 1),
 (8, 'Lucas Dubois',    'lucas.dubois@example.com',    'FR',     '2023-05-12', 1),
 (9, 'Mia Andersen',    'mia.andersen@example.com',    'DK',     '2023-05-30', 0),
(10, 'Ethan Brown',     'ethan.brown@example.com',     'US',     '2023-06-15', 1),
(11, 'Olivia Wang',     'olivia.wang@example.com',     'CN',     '2023-07-02', 1),
(12, 'Henry Schmidt',   'henry.schmidt@example.com',   'DE',     '2023-07-21', 1),
(13, 'Aria Lopez',      'aria.lopez@example.com',      'MX',     '2023-08-04', 1),
(14, 'Oscar Lindgren',  'oscar.lindgren@example.com',  'SE',     '2023-08-19', 1),
(15, 'Chloe Nakamura',  'chloe.nakamura@example.com',  'JP',     '2023-09-07', 1),
(16, 'Finn O''Connor',  'finn.oconnor@example.com',    'IE',     '2023-09-25', 1),
(17, 'Layla Hassan',    'layla.hassan@example.com',    'EG',     '2023-10-10', 1),
(18, 'Caleb Anderson',  'caleb.anderson@example.com',  'US',     '2023-10-28', 0),
(19, 'Nora Jensen',     'nora.jensen@example.com',     'NO',     '2023-11-13', 1),
(20, 'Diego Silva',     'diego.silva@example.com',     'BR',     '2023-11-29', 1),
(21, 'Emily Carter',    'emily.carter@example.com',    'UK',     '2024-01-04', 1),
(22, 'Hiroshi Tanaka',  'hiroshi.tanaka@example.com',  'JP',     '2024-01-22', 1),
(23, 'Priya Sharma',    'priya.sharma@example.com',    'IN',     '2024-02-08', 1),
(24, 'Lukas Becker',    'lukas.becker@example.com',    'DE',     '2024-02-26', 1),
(25, 'Amelia Wright',   'amelia.wright@example.com',   'UK',     '2024-03-15', 1),
(26, 'Tomás Fernández', 'tomas.fernandez@example.com', 'ES',     '2024-04-02', 1),
(27, 'Hannah Moller',   'hannah.moller@example.com',   'AT',     '2024-04-19', 1),
(28, 'Jin Park',        'jin.park@example.com',        'KR',     '2024-05-05', 1),
(29, 'Beatrice Conti',  'beatrice.conti@example.com',  'IT',     '2024-05-23', 1),
(30, 'Daniel Mwangi',   'daniel.mwangi@example.com',   'KE',     '2024-06-10', 1),
(31, 'Sophie Laurent',  'sophie.laurent@example.com',  'FR',     '2024-07-01', 1),
(32, 'Ben Walker',      'ben.walker@example.com',      'AU',     '2024-07-18', 1),
(33, 'Ruby Bennett',    'ruby.bennett@example.com',    'UK',     '2024-08-06', 1),
(34, 'Adrián Vega',     'adrian.vega@example.com',     'CL',     '2024-08-24', 1),
(35, 'Maya Cohen',      'maya.cohen@example.com',      'IL',     '2024-09-09', 1),
(36, 'Theo Holmes',     'theo.holmes@example.com',     'UK',     '2024-09-27', 1),
(37, 'Sienna Ng',       'sienna.ng@example.com',       'SG',     '2024-10-14', 1),
(38, 'Marcus Reid',     'marcus.reid@example.com',     'CA',     '2024-11-02', 1),
(39, 'Eva Novak',       'eva.novak@example.com',       'CZ',     '2024-11-20', 1),
(40, 'Yuki Sato',       'yuki.sato@example.com',       'JP',     '2024-12-08', 1);

-- ============================================================
-- Products (30) with JSON metadata
-- ============================================================
INSERT INTO products (id, sku, name, category, price, stock, metadata, created_at) VALUES
 (1,  'KBD-001', 'Mechanical Keyboard 75%',     'peripherals', 149.99, 42,  '{"layout":"75%","switches":"brown","backlight":"RGB","weight_g":820}', '2024-03-12 09:15:00'),
 (2,  'KBD-002', 'Ergonomic Split Keyboard',    'peripherals', 229.00, 17,  '{"layout":"split","switches":"silent-red","wireless":true,"battery_mah":4000}', '2024-03-18 11:42:00'),
 (3,  'MSE-001', 'Wireless Gaming Mouse',       'peripherals',  79.50, 88,  '{"dpi":26000,"buttons":8,"weight_g":63,"wireless":true}', '2024-03-21 14:08:00'),
 (4,  'MSE-002', 'Trackball Mouse',             'peripherals',  64.00, 31,  '{"type":"trackball","hand":"right","wireless":false}', '2024-04-02 10:30:00'),
 (5,  'MON-001', '27" 4K IPS Monitor',          'displays',    389.00, 23,  '{"size_in":27,"resolution":"3840x2160","panel":"IPS","hdr":true,"refresh_hz":60}', '2024-04-15 16:22:00'),
 (6,  'MON-002', '34" Ultrawide Curved',        'displays',    649.00,  9,  '{"size_in":34,"resolution":"3440x1440","panel":"VA","hdr":true,"refresh_hz":144,"curve":"1500R"}', '2024-04-20 09:05:00'),
 (7,  'CHR-001', 'Ergonomic Mesh Chair',        'furniture',   459.00, 12,  '{"colour":"charcoal","material":"mesh","warranty_years":5,"max_load_kg":136}', '2024-05-04 13:18:00'),
 (8,  'DSK-001', 'Standing Desk Electric',      'furniture',   599.00,  6,  '{"width_cm":160,"depth_cm":80,"height_range_cm":[65,130],"motors":2}', '2024-05-11 11:00:00'),
 (9,  'HDP-001', 'Noise-Cancelling Headphones', 'audio',       329.00, 54,  '{"driver_mm":40,"battery_hours":30,"anc":true,"bluetooth":"5.3","codecs":["LDAC","aptX"]}', '2024-05-25 15:45:00'),
(10, 'HDP-002', 'Studio Reference Headphones',  'audio',       199.00, 28,  '{"driver_mm":50,"impedance_ohm":250,"open_back":true}', '2024-06-02 08:20:00'),
(11, 'SPK-001', 'Bluetooth Bookshelf Speakers', 'audio',       279.00, 19,  '{"wattage":80,"bluetooth":"5.0","subwoofer_out":true,"colour":"walnut"}', '2024-06-14 12:10:00'),
(12, 'WCM-001', '4K Streaming Webcam',          'peripherals', 159.00, 37,  '{"resolution":"4K","fov_deg":90,"autofocus":true,"hdr":true}', '2024-06-22 17:55:00'),
(13, 'MIC-001', 'USB Condenser Microphone',    'audio',       129.00, 44,  '{"pattern":["cardioid","omni","bidirectional"],"sample_rate_khz":48,"connector":"USB-C"}', '2024-07-03 10:40:00'),
(14, 'NAS-001', '4-Bay Network Attached Storage','networking', 549.00,  8,  '{"bays":4,"cpu":"ARM Cortex-A57","ram_gb":4,"raid":[0,1,5,6,10]}', '2024-07-09 14:25:00'),
(15, 'SSD-001', '2TB NVMe SSD Gen4',           'storage',     189.00, 73,  '{"capacity_tb":2,"interface":"PCIe 4.0","read_mbps":7000,"write_mbps":6500}', '2024-07-17 09:30:00'),
(16, 'SSD-002', '4TB External SSD USB-C',      'storage',     349.00, 25,  '{"capacity_tb":4,"interface":"USB-C 3.2","read_mbps":1050}', '2024-07-22 11:50:00'),
(17, 'HUB-001', 'USB-C Docking Station',       'accessories',  89.00, 61,  '{"ports":{"hdmi":2,"usb_a":3,"usb_c":2,"sd":1,"ethernet":1},"pd_w":100}', '2024-08-01 13:00:00'),
(18, 'CBL-001', 'Thunderbolt 4 Cable 2m',      'accessories',  49.00,150,  '{"length_m":2,"speed_gbps":40,"pd_w":100}', '2024-08-06 09:12:00'),
(19, 'PWR-001', 'USB-C 140W Charger',          'accessories',  79.00, 92,  '{"wattage":140,"ports":3,"gan":true}', '2024-08-13 15:33:00'),
(20, 'LMP-001', 'Monitor Light Bar',           'accessories',  119.00, 47, '{"width_cm":50,"colour_temp_k":[2700,6500],"auto_dim":true}', '2024-08-21 10:18:00'),
(21, 'CHR-002', 'Gaming Racing Chair',         'furniture',   349.00, 14,  '{"colour":"red-black","material":"PU leather","reclining":true}', '2024-09-04 12:42:00'),
(22, 'DSK-002', 'Walnut Wooden Desk',          'furniture',   429.00, 11,  '{"width_cm":140,"depth_cm":70,"material":"solid walnut"}', '2024-09-12 14:50:00'),
(23, 'TAB-001', '11" Pro Tablet',              'tablets',     799.00, 22,  '{"screen_in":11,"storage_gb":256,"ram_gb":8,"stylus":true}', '2024-09-25 16:08:00'),
(24, 'LAP-001', '14" Ultrabook',               'laptops',    1299.00, 16,  '{"screen_in":14,"cpu":"snapdragon-x","ram_gb":16,"storage_gb":512,"weight_kg":1.2}', '2024-10-02 11:30:00'),
(25, 'LAP-002', '16" Pro Workstation',         'laptops',    2499.00,  7,  '{"screen_in":16,"cpu":"M3 Pro","ram_gb":36,"storage_gb":1024,"weight_kg":2.1}', '2024-10-11 09:55:00'),
(26, 'PHN-001', '6.7" Flagship Phone',         'phones',      999.00, 33,  '{"screen_in":6.7,"storage_gb":256,"camera_mp":50,"5g":true}', '2024-10-22 13:14:00'),
(27, 'WCH-001', 'Smart Watch Series 9',        'wearables',   429.00, 41,  '{"display":"AMOLED","gps":true,"ecg":true,"battery_hours":18}', '2024-11-03 10:00:00'),
(28, 'WCH-002', 'Sports GPS Watch',            'wearables',   249.00, 29,  '{"battery_days":14,"gps":"dual-band","water_resistance_m":50}', '2024-11-15 15:20:00'),
(29, 'RTR-001', 'Wi-Fi 7 Mesh Router (3-pack)','networking',  599.00, 13,  '{"standard":"Wi-Fi 7","coverage_sqm":650,"nodes":3,"ports":4}', '2024-11-28 11:45:00'),
(30, 'GPU-001', 'Graphics Card 16GB',          'components',  799.00, 18,  '{"memory_gb":16,"interface":"PCIe 4.0","tdp_w":250,"hdmi":1,"dp":3}', '2024-12-05 14:33:00');

-- ============================================================
-- Orders (60) — varied statuses and dates
-- ============================================================
INSERT INTO orders (id, customer_id, status, total, currency, created_at, notes) VALUES
 (1001,  1, 'delivered', 229.49, 'GBP', '2024-09-04 10:14:00', NULL),
 (1002,  3, 'delivered', 389.00, 'EUR', '2024-09-06 14:32:00', NULL),
 (1003,  2, 'delivered',  79.50, 'GBP', '2024-09-09 09:18:00', 'Gift wrap requested'),
 (1004,  5, 'delivered', 459.00, 'GBP', '2024-09-12 11:45:00', NULL),
 (1005,  7, 'delivered', 198.50, 'GBP', '2024-09-14 16:02:00', NULL),
 (1006,  8, 'delivered', 329.00, 'EUR', '2024-09-15 08:50:00', NULL),
 (1007, 10, 'delivered', 649.00, 'USD', '2024-09-18 19:21:00', NULL),
 (1008, 11, 'delivered', 189.00, 'USD', '2024-09-22 12:08:00', NULL),
 (1009, 12, 'delivered', 119.00, 'EUR', '2024-09-25 10:30:00', NULL),
 (1010, 15, 'delivered', 799.00, 'JPY', '2024-09-28 22:14:00', 'Express delivery'),
 (1011,  4, 'delivered', 279.00, 'KRW', '2024-10-01 11:00:00', NULL),
 (1012, 14, 'delivered', 559.00, 'SEK', '2024-10-04 14:20:00', NULL),
 (1013, 19, 'delivered', 349.00, 'NOK', '2024-10-07 09:42:00', NULL),
 (1014, 22, 'delivered', 999.00, 'JPY', '2024-10-10 18:35:00', NULL),
 (1015, 21, 'delivered',  89.00, 'GBP', '2024-10-12 13:08:00', NULL),
 (1016, 23, 'delivered',1299.00, 'INR', '2024-10-15 11:55:00', NULL),
 (1017, 25, 'delivered', 429.00, 'GBP', '2024-10-18 16:40:00', NULL),
 (1018, 27, 'delivered', 249.00, 'EUR', '2024-10-21 10:12:00', NULL),
 (1019, 30, 'delivered', 149.99, 'KES', '2024-10-24 09:33:00', NULL),
 (1020, 31, 'delivered', 599.00, 'EUR', '2024-10-27 15:50:00', NULL),
 (1021,  6, 'delivered', 199.00, 'EUR', '2024-10-30 12:18:00', NULL),
 (1022, 13, 'delivered', 159.00, 'MXN', '2024-11-02 11:05:00', NULL),
 (1023, 16, 'delivered', 549.00, 'EUR', '2024-11-05 17:24:00', NULL),
 (1024, 17, 'delivered',  79.00, 'EGP', '2024-11-08 10:00:00', NULL),
 (1025, 20, 'delivered', 129.00, 'BRL', '2024-11-11 13:48:00', NULL),
 (1026, 28, 'delivered', 349.00, 'KRW', '2024-11-14 19:10:00', NULL),
 (1027, 32, 'delivered', 429.00, 'AUD', '2024-11-17 09:22:00', NULL),
 (1028, 33, 'delivered', 599.00, 'GBP', '2024-11-20 14:15:00', NULL),
 (1029, 34, 'delivered',2499.00, 'CLP', '2024-11-23 11:02:00', 'Insurance added'),
 (1030, 35, 'delivered', 799.00, 'ILS', '2024-11-26 16:30:00', NULL),
 (1031,  1, 'shipped',   119.00, 'GBP', '2025-01-08 10:11:00', NULL),
 (1032,  5, 'shipped',   229.00, 'GBP', '2025-01-11 14:45:00', NULL),
 (1033, 21, 'shipped',   249.00, 'GBP', '2025-01-14 12:02:00', NULL),
 (1034, 25, 'shipped',   189.00, 'GBP', '2025-01-17 09:38:00', 'Leave with neighbour if out'),
 (1035, 33, 'shipped',   349.00, 'GBP', '2025-01-20 11:24:00', NULL),
 (1036, 36, 'shipped',   429.00, 'GBP', '2025-01-23 15:18:00', NULL),
 (1037,  7, 'shipped',   299.50, 'GBP', '2025-01-26 13:45:00', NULL),
 (1038, 37, 'shipped',   799.00, 'SGD', '2025-01-29 10:09:00', NULL),
 (1039, 38, 'shipped',  1299.00, 'CAD', '2025-02-01 16:55:00', NULL),
 (1040, 39, 'shipped',   159.00, 'CZK', '2025-02-04 11:32:00', NULL),
 (1041,  2, 'paid',      149.99, 'GBP', '2025-02-08 09:14:00', NULL),
 (1042,  3, 'paid',      649.00, 'EUR', '2025-02-10 14:48:00', NULL),
 (1043, 10, 'paid',      329.00, 'USD', '2025-02-12 11:05:00', NULL),
 (1044, 22, 'paid',      999.00, 'JPY', '2025-02-14 17:22:00', NULL),
 (1045, 40, 'paid',      599.00, 'JPY', '2025-02-15 10:30:00', NULL),
 (1046,  4, 'paid',      249.00, 'KRW', '2025-02-17 13:12:00', NULL),
 (1047, 15, 'paid',      429.00, 'JPY', '2025-02-19 16:00:00', NULL),
 (1048, 31, 'paid',      199.00, 'EUR', '2025-02-21 09:48:00', NULL),
 (1049, 26, 'pending',    79.00, 'EUR', '2025-02-22 11:30:00', NULL),
 (1050, 12, 'pending',   549.00, 'EUR', '2025-02-23 14:22:00', NULL),
 (1051, 27, 'pending',   119.00, 'EUR', '2025-02-24 10:15:00', NULL),
 (1052, 32, 'pending',   349.00, 'AUD', '2025-02-25 12:40:00', NULL),
 (1053, 17, 'cancelled', 159.00, 'EGP', '2024-12-03 09:12:00', 'Customer changed mind'),
 (1054,  9, 'cancelled', 229.00, 'DKK', '2024-12-08 13:48:00', 'Out of stock at fulfilment'),
 (1055, 18, 'cancelled', 599.00, 'USD', '2024-12-15 11:25:00', NULL),
 (1056,  6, 'refunded',  329.00, 'EUR', '2024-11-29 14:05:00', 'Defective on arrival'),
 (1057, 24, 'refunded',  189.00, 'EUR', '2024-12-12 16:18:00', 'Wrong item shipped'),
 (1058, 29, 'refunded',  799.00, 'EUR', '2025-01-04 10:42:00', 'Damaged in transit'),
 (1059, 11, 'delivered', 459.00, 'CNY', '2024-12-20 12:00:00', NULL),
 (1060, 19, 'delivered', 129.00, 'NOK', '2024-12-27 15:33:00', NULL);

-- ============================================================
-- Order items (one or more per order)
-- ============================================================
INSERT INTO order_items (order_id, product_id, quantity, unit_price) VALUES
 (1001,  2, 1, 229.00),
 (1001, 18, 1,   0.49),
 (1002,  5, 1, 389.00),
 (1003,  3, 1,  79.50),
 (1004,  7, 1, 459.00),
 (1005,  1, 1, 149.99),
 (1005, 18, 1,  48.51),
 (1006,  9, 1, 329.00),
 (1007,  6, 1, 649.00),
 (1008, 15, 1, 189.00),
 (1009, 20, 1, 119.00),
 (1010, 23, 1, 799.00),
 (1011, 11, 1, 279.00),
 (1012,  9, 1, 329.00),
 (1012, 19, 1, 230.00),
 (1013, 21, 1, 349.00),
 (1014, 26, 1, 999.00),
 (1015, 17, 1,  89.00),
 (1016, 24, 1,1299.00),
 (1017, 27, 1, 429.00),
 (1018, 28, 1, 249.00),
 (1019,  1, 1, 149.99),
 (1020, 29, 1, 599.00),
 (1021, 10, 1, 199.00),
 (1022, 12, 1, 159.00),
 (1023, 14, 1, 549.00),
 (1024, 19, 1,  79.00),
 (1025, 13, 1, 129.00),
 (1026, 21, 1, 349.00),
 (1027, 27, 1, 429.00),
 (1028, 29, 1, 599.00),
 (1029, 25, 1,2499.00),
 (1030, 23, 1, 799.00),
 (1031, 20, 1, 119.00),
 (1032,  2, 1, 229.00),
 (1033, 28, 1, 249.00),
 (1034, 15, 1, 189.00),
 (1035, 21, 1, 349.00),
 (1036, 27, 1, 429.00),
 (1037,  9, 1, 299.50),
 (1038, 23, 1, 799.00),
 (1039, 24, 1,1299.00),
 (1040, 12, 1, 159.00),
 (1041,  1, 1, 149.99),
 (1042,  6, 1, 649.00),
 (1043,  9, 1, 329.00),
 (1044, 26, 1, 999.00),
 (1045, 29, 1, 599.00),
 (1046, 28, 1, 249.00),
 (1047, 27, 1, 429.00),
 (1048, 10, 1, 199.00),
 (1049, 19, 1,  79.00),
 (1050, 14, 1, 549.00),
 (1051, 20, 1, 119.00),
 (1052, 21, 1, 349.00),
 (1053, 12, 1, 159.00),
 (1054,  2, 1, 229.00),
 (1055, 29, 1, 599.00),
 (1056,  9, 1, 329.00),
 (1057, 15, 1, 189.00),
 (1058, 23, 1, 799.00),
 (1059,  7, 1, 459.00),
 (1060, 13, 1, 129.00);

-- ============================================================
-- Reviews (mixed ratings, varied length text for cell popover)
-- ============================================================
INSERT INTO reviews (product_id, customer_id, rating, title, body, created_at) VALUES
 (1,  1, 5, 'Love the typing feel',          'The brown switches are tactile without being noisy — perfect for shared offices. RGB is more useful than I expected for finding the function row in low light.', '2024-09-20 19:14:00'),
 (1,  5, 4, 'Solid but heavy',               'Build quality is excellent and the typing experience is fantastic. Only knocking a star because it''s too heavy to comfortably travel with.',                                          '2024-10-02 11:30:00'),
 (2,  7, 5, 'Saved my wrists',               'After years of pain I switched to this split layout and the difference within two weeks was dramatic. The wireless range is genuinely solid and battery lasts about three weeks of heavy use.', '2024-10-14 09:48:00'),
 (3,  2, 5, 'Light and fast',                'At 63g it disappears in your hand. Tracking is flawless on glass and the side buttons are perfectly placed.',                                                                                  '2024-09-22 14:05:00'),
 (5,  3, 4, 'Great panel, weak stand',       'Colours are gorgeous out of the box and HDR is a real upgrade. Stand wobbles a little — replaced with a VESA arm and it''s perfect.',                                                       '2024-09-25 16:18:00'),
 (6, 10, 5, 'Ultrawide changed my workflow', 'Going back to a 16:9 monitor feels broken now. The 144Hz refresh is silky for both work and gaming.',                                                                                       '2024-10-05 21:02:00'),
 (7,  5, 5, 'Worth every penny',             'I''ve owned three "ergonomic" chairs in the last decade and this is the first one I''d actually call ergonomic. The lumbar support is the best part.',                                     '2024-10-18 13:25:00'),
 (9,  8, 5, 'ANC is incredible',             'Used these on a long-haul flight and forgot the engines were even there. Bluetooth pairing has been rock solid across phone and laptop.',                                                  '2024-10-08 22:40:00'),
 (9,  6, 2, 'Pads wear out fast',            'Sound is genuinely superb but the ear pads started flaking after about four months. Replacement pads aren''t cheap.',                                                                       '2024-11-30 09:14:00'),
(10,  4, 4, 'Honest reference sound',        'These are not "fun" headphones — they''re honest. Mixing on them translates well to other systems which is the whole point.',                                                              '2024-10-25 18:08:00'),
(15, 11, 5, 'Blazing fast',                  'Cloned a 1TB game library in under three minutes. No throttling under sustained writes either.',                                                                                            '2024-09-30 11:42:00'),
(15, 22, 4, 'Runs warm',                     'Excellent performance but it does get noticeably warm under sustained load. A heatsink is a must if you plan heavy use.',                                                                  '2024-10-30 15:10:00'),
(17, 21, 5, 'Cleaned up my desk',            'Replaces four separate adapters. The 100W passthrough easily charges my work laptop while everything else is connected.',                                                                  '2024-10-22 10:05:00'),
(23, 15, 5, 'Pro tablet for the price',      'The screen is honestly indistinguishable from far more expensive options. Stylus latency is barely perceptible.',                                                                          '2024-10-06 17:30:00'),
(24, 23, 5, 'Perfect ultrabook',             'Battery genuinely lasts a full workday with brightness up. Fanless under most workloads.',                                                                                                 '2024-10-22 09:18:00'),
(25, 34, 5, 'Workstation in a backpack',     'Compiles that took 7 minutes on my old machine finish in under 90 seconds. Worth the upgrade.',                                                                                            '2024-11-30 14:48:00'),
(26, 22, 4, 'Great phone, mediocre camera',  'Performance and battery are excellent but low-light photos still lag behind the obvious competitor. Day-to-day I love it.',                                                                '2024-10-15 19:12:00'),
(27, 25, 5, 'ECG actually works',            'Detected an arrhythmia my GP missed for years. Daily comfort is great and battery hits the advertised 18 hours.',                                                                          '2024-10-25 08:55:00'),
(29, 31, 4, 'Wi-Fi 7 is real',               'Coverage in my 4-bedroom flat is now flawless. App could be better but performance is what matters.',                                                                                       '2024-11-08 12:30:00'),
(30, NULL, 3, NULL,                          NULL,                                                                                                                                                                                        '2024-12-10 11:00:00');

-- ============================================================
-- Quick sanity check
-- ============================================================
SELECT 'customers'    AS table_name, COUNT(*) AS rows FROM customers
UNION ALL SELECT 'products',    COUNT(*) FROM products
UNION ALL SELECT 'orders',      COUNT(*) FROM orders
UNION ALL SELECT 'order_items', COUNT(*) FROM order_items
UNION ALL SELECT 'reviews',     COUNT(*) FROM reviews;
