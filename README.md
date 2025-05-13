# 🧀 Kejucraft

_Tubes 2 Strategi Algoritma - IF2211_

Kejucraft adalah sebuah website yang memungkinkan pengguna mencari jalur crafting dari elemen-elemen dasar menuju elemen target berdasarkan data dari _game_ **Little Alchemy 2**. Aplikasi ini memanfaatkan algoritma pencarian seperti **BFS**, **DFS**, **Bidirectional BFS**, dan **Bidirectional DFS**, serta menyediakan dua mode pencarian: **Single Recipe** (satu jalur crafting tercepat) dan **Multiple Recipes** (beragam jalur crafting unik).

Seluruh data elemen dan kombinasi crafting diambil langsung melalui proses scraping dari halaman [Little Alchemy 2 Wiki](https://little-alchemy.fandom.com/wiki/Elements_(Little_Alchemy_2)).

---

## 🔍 Fitur Utama

- 🔗 **Web scraping otomatis** elemen dan kombinasi resep dari wiki Little Alchemy 2.
- 🧠 **Pilihan algoritma pencarian**:
  - BFS (Breadth-First Search)
  - DFS (Depth-First Search)
  - Bidirectional BFS
  - Bidirectional DFS
- 🧪 **Dua mode pencarian:**
  - Single Recipe: mencari jalur crafting paling efisien
  - Multiple Recipes: menghasilkan variasi jalur crafting unik

---

## Dependensi
### Node Package Manager 
```bash
cd fe
npm install
npm run dev
```
### Golang
Kunjungi situs resmi Go dan sesuaikan dengan versi OS:

[https://go.dev/dl](https://go.dev/dl)

---

## Cara Menjalankan Aplikasi

### Melalui Website 
ttps://your-deploy-url-here.com

### Melalui Docker

# 1. Buat terlebih dahulu network docker
```bash
docker network create kejucraft-net
```
# 2. Jalankan backend
```
cd be
docker compose up
```

# 3. Jalankan frontend (di terminal baru)
```
cd ../fe
docker compose up
```

### Melalui Local Terminal

# 1. Clone dua repository berikut dan tempatkan dalam satu folder
```
git clone https://github.com/sbimasena/Tubes2_FE_bagikopiciumbuleuit.git ke dalam folder fe
git clone https://github.com/sbimasena/Tubes2_BE_bagikopiciumbuleuit.git ke dalam folder be
```
# 2. Jalankan Frontend (React + Next.js)
```
cd fe/kejucraft
npm install
npm run dev
```

# 3. Jalankan Backend (Golang)
```
cd ../be
go run *.go
```

## Kontributor

| NIM      | Nama                  |
| -------- | --------------------- |
| 13523049 | Muhammad Fithra Rizki |
| 13523053 | Sakti Bimasena        |
| 13523062 | Aliya Husna Fayyaza   |


