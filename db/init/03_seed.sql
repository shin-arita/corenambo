INSERT INTO products (sku, name, description, price, is_active)
VALUES
('SKU-0001', '防水シューズ', '日本語の商品説明です。軽量で防水性があります。雨の日の外出や通勤にも使いやすい商品です。', 9800, true),
('SKU-0002', 'ランニングウェア', '通気性が高く、夏場のランニングに向いています。軽量素材を採用し、着心地も快適です。', 5400, true),
('SKU-0003', '保温ボトル', '長時間保温できるステンレスボトルです。日常利用にもアウトドアにも適しています。', 3200, true)
ON CONFLICT (sku) DO NOTHING;
