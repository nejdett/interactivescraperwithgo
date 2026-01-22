# CTI Veri Toplama Sistemi

Siber tehdit istihbaratı toplama ve analiz platformu.

## Ne İşe Yarar?

Bu proje dark web forumlarından ve çeşitli kaynaklardan siber tehdit verilerini otomatik olarak toplar. Tor ağı üzerinden .onion sitelerine bağlanabilir, içerikleri kategorize eder ve tehdit seviyelerini değerlendirir.

### Özellikler

- Tor entegrasyonu ile dark web erişimi
- Forum ve RSS feed tarama
- Otomatik başlık üretimi
- Kritiklik skorlama (1-10)
- Otomatik kategorilendirme
- Web dashboard
- Periyodik veri toplama

## Teknolojiler

**Backend:** Go, PostgreSQL, Docker
**Network:** Tor Proxy (SOCKS5)
**Dashboard:** Go templates, Gorilla sessions

## Kurulum

### Gereksinimler

- Docker & Docker Compose
- En az 2GB RAM

### Adımlar

1. Projeyi klonlayın
```bash
git clone <repo-url>
cd interactive_scraper
```

2. `.env` dosyasını düzenleyin
```bash
DB_PASSWORD=güvenli_şifre
SESSION_SECRET=en_az_32_karakter
COLLECTION_INTERVAL=5m
USE_TOR=true
```

3. Başlatın
```bash
docker-compose up -d
```

4. Dashboard'a gidin: `http://localhost:8080`

### Giriş Bilgileri

Sistem ilk kurulumda otomatik admin kullanıcısı oluşturur:

**Kullanıcı:** admin  
**Şifre:** admin123

> İlk girişten sonra şifrenizi değiştirin!

### Logları görüntüleme

```bash
docker-compose logs -f
docker logs -f cti-collector
```

### Durdurma

```bash
docker-compose down
# verileri de silmek için:
docker-compose down -v
```

## Başlık Üretimi

Sistem içeriklerden otomatik başlık üretir:

1. İçerik temizlenir (boşluklar, özel karakterler)
2. İlk cümle çıkarılır
3. Max 100 karakter ile sınırlanır
4. Kelime ortasında kesilmez

Eğer başlık çok kısa çıkarsa (<10 karakter), içeriğin ilk 100 karakteri kullanılır.

**Örnek:**
```
İçerik: "New vulnerability discovered in Apache Log4j allows remote code execution..."
Başlık: "New vulnerability discovered in Apache Log4j allows remote code execution"
```

## Kritiklik Skorlama

İçerikler 1-10 arası skorlanır:

- **9-10:** Kritik (0-day, RCE, data breach)
- **7-8:** Yüksek (exploit, ransomware)
- **5-6:** Orta (vulnerability, phishing)
- **3-4:** Düşük (genel tartışma)
- **1-2:** Bilgi (haberler)

Skorlama anahtar kelime analizi ve kategori eşleştirmesi ile yapılır.

## Proje Yapısı

```
interactive_scraper/
├── collector/              # veri toplama servisi
│   ├── cmd/collector/     # ana uygulama
│   ├── internal/
│   │   ├── scraper/       # scraping modülleri
│   │   ├── processor/     # içerik işleme
│   │   └── repository/    # database işlemleri
│   └── sources.json       # veri kaynakları
├── dashboard/             # web arayüzü
│   ├── cmd/dashboard/
│   ├── internal/
│   │   ├── handlers/
│   │   └── repository/
│   └── web/templates/
├── database/init.sql      # db şeması
├── docker-compose.yml
└── .env
```

## Yapılandırma

### Kaynak Ekleme

`collector/sources.json` dosyasını düzenleyin:
```json
[
    {
        "name": "Forum Adı",
        "url": "http://example.onion/"
    }
]
```

### Toplama Aralığı

`.env` dosyasında:
```
COLLECTION_INTERVAL=5m   # 5 dakika
COLLECTION_INTERVAL=1h   # 1 saat
```

### Log Seviyesi

```
LOG_LEVEL=debug  # detaylı
LOG_LEVEL=info   # normal (önerilen)
LOG_LEVEL=error  # sadece hatalar
```

## Veritabanı

### Tablolar

**content_items** - Ana içerik tablosu
- id, title, source_name, source_url
- content, published_at
- criticality_score (1-10)
- created_at, updated_at

**categories** - Kategori tanımları
- id, name, description

**content_categories** - İçerik-kategori ilişkisi

## Sorun Giderme

### Tor bağlanamıyor

```bash
docker logs cti-tor
docker-compose restart tor
```

### Database hatası

```bash
docker logs cti-postgres
docker exec -it cti-postgres psql -U cti_user -d cti_db
```

### Collector çalışmıyor

```bash
docker logs cti-collector --tail 100
docker-compose restart collector
```

## İpuçları

- Toplama aralığını çok kısa tutmayın (min 5dk önerilir)
- Aynı anda çok fazla kaynak taramayın
- Eski verileri düzenli temizleyin
- Tor yavaş olabilir, timeout'ları artırın

## Güvenlik

- `.env` dosyasını paylaşmayın
- Güçlü şifreler kullanın
- Production'da reverse proxy kullanın
- Docker image'ları güncel tutun

## Lisans

Eğitim ve araştırma amaçlıdır. Yasal olmayan kullanım yasaktır.

## Katkı

1. Fork yapın
2. Feature branch oluşturun
3. Commit edin
4. Push edin
5. Pull request açın

---

**Uyarı:** Bu sistem dark web içeriklerine erişir. Yasal ve etik kullanım sorumluluğu kullanıcıya aittir.
