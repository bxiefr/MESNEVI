package main

import (
	"encoding/json"
	"fmt"
	"os"
	"syscall"
	"time"
	"unsafe"
)

// WinAPI DLL ve Fonksiyon Tanımlamaları
var (
	user32   = syscall.NewLazyDLL("user32.dll")
	kernel32 = syscall.NewLazyDLL("kernel32.dll")
	gdi32    = syscall.NewLazyDLL("gdi32.dll")
	winmm    = syscall.NewLazyDLL("winmm.dll")

	procOpenProcess       = kernel32.NewProc("OpenProcess")
	procReadProcessMemory = kernel32.NewProc("ReadProcessMemory")
	procCloseHandle       = kernel32.NewProc("CloseHandle")
	procPlaySoundW        = winmm.NewProc("PlaySoundW")
	procGetSystemMetrics  = user32.NewProc("GetSystemMetrics")

	// GDI Çizim Fonksiyonları
	procCreateCompatibleDC     = gdi32.NewProc("CreateCompatibleDC")
	procCreateCompatibleBitmap = gdi32.NewProc("CreateCompatibleBitmap")
	procSelectObject           = gdi32.NewProc("SelectObject")
	procDeleteObject           = gdi32.NewProc("DeleteObject")
	procDeleteDC               = gdi32.NewProc("DeleteDC")
	procCreatePen              = gdi32.NewProc("CreatePen")
	procMoveToEx               = gdi32.NewProc("MoveToEx")
	procLineTo                 = gdi32.NewProc("LineTo")
	procBitBlt                 = gdi32.NewProc("BitBlt")
)

const (
	PROCESS_VM_READ           = 0x0010
	PROCESS_VM_OPERATION      = 0x0008
	SND_FILENAME              = 0x00020000
	SND_ASYNC                 = 0x0001
	PS_SOLID                  = 0
	SRCCOPY                   = 0x00CC0020
)

// Offset Yapısı (Verdiğin JSON ile Tam Uyumlu)
type Offsets struct {
	DwViewMatrix           uintptr `json:"dwViewMatrix"`
	DwLocalPlayerPawn      uintptr `json:"dwLocalPlayerPawn"`
	DwEntityList           uintptr `json:"dwEntityList"`
	M_hPlayerPawn          uintptr `json:"m_hPlayerPawn"`
	M_iHealth              uintptr `json:"m_iHealth"`
	M_lifeState            uintptr `json:"m_lifeState"`
	M_iTeamNum             uintptr `json:"m_iTeamNum"`
	M_vOldOrigin           uintptr `json:"m_vOldOrigin"`
	M_pGameSceneNode       uintptr `json:"m_pGameSceneNode"`
	M_modelState           uintptr `json:"m_modelState"`
	M_boneArray            uintptr `json:"m_boneArray"`
	M_nodeToWorld          uintptr `json:"m_nodeToWorld"`
	M_sSanitizedPlayerName uintptr `json:"m_sSanitizedPlayerName"`
	M_ArmorValue           uintptr `json:"m_ArmorValue"`
	M_pBulletServices      uintptr `json:"m_pBulletServices"`
	M_totalHitsOnServer    uintptr `json:"m_totalHitsOnServer"`
}

var (
	offsets    Offsets
	gameHandle uintptr
	clientDLL  uintptr

	// Hitmarker Durumu
	lastTotalHits   int32     = -1
	hitmarkerActive bool      = false
	hitmarkerTime   time.Time
)

// Generic Bellek Okuma Fonksiyonu
func ReadMemory[T any](address uintptr) T {
	var result T
	if address == 0 || gameHandle == 0 {
		return result
	}
	procReadProcessMemory.Call(
		gameHandle,
		address,
		uintptr(unsafe.Pointer(&result)),
		unsafe.Sizeof(result),
		0,
	)
	return result
}

// Asenkron Ses Çalma Fonksiyonu
func playHitSound() {
	// Örnek: Çalışma dizinindeki "hitsound.wav" dosyasını çalar.
	// Dosya yoksa Windows varsayılan bip sesini kullanabilir veya SND_ALIAS verebilirsiniz.
	soundPath, _ := syscall.UTF16PtrFromString("hitsound.wav")
	procPlaySoundW.Call(
		uintptr(unsafe.Pointer(soundPath)),
		0,
		uintptr(SND_FILENAME|SND_ASYNC),
	)
}

// 1. Sadece Senin Mermilerinle Tetiklenen Hitmarker Kontrolü
func UpdateHitmarker(localPawn uintptr) {
	if localPawn == 0 {
		lastTotalHits = -1
		return
	}

	// Local player'ın CPlayer_BulletServices yapısının adresini oku
	bulletServices := ReadMemory[uintptr](localPawn + offsets.M_pBulletServices)
	if bulletServices == 0 {
		return
	}

	// Sunucunun onayladığı isabet sayısını oku (m_totalHitsOnServer)
	currentHits := ReadMemory[int32](bulletServices + offsets.M_totalHitsOnServer)

	// Oyuna ilk girişte veya öldükten sonra doğuşta değeri eşitle
	if lastTotalHits == -1 {
		lastTotalHits = currentHits
		return
	}

	// Eğer sunucudaki isabet sayısı arttıysa kesinlikle senin attığın mermi vurmuştur
	if currentHits > lastTotalHits {
		hitmarkerActive = true
		hitmarkerTime = time.Now()

		// Oyunu dondurmamak için sesi ayrı bir goroutine'de çalıştır
		go playHitSound()

		lastTotalHits = currentHits
	}
}

// 2. Hitmarker Çizim Fonksiyonu (GDI Double Buffering Destekli)
func DrawHitmarker(memDC uintptr, screenWidth, screenHeight int32) {
	if !hitmarkerActive {
		return
	}

	// 300ms sonra ekrandan sil
	if time.Since(hitmarkerTime) > 300*time.Millisecond {
		hitmarkerActive = false
		return
	}

	centerX := screenWidth / 2
	centerY := screenHeight / 2
	size := int32(7) // X çizgisinin boyutu (Piksel)

	// Beyaz/Kırmızı Kalem Oluştur (RGB: 255, 0, 0 -> Kırmızı)
	pen, _, _ := procCreatePen.Call(PS_SOLID, 2, uintptr(0x0000FF)) // RGB(255, 0, 0) - Kırmızı
	oldPen, _, _ := procSelectObject.Call(memDC, pen)

	// Sol Üst -> Sağ Alt Çizgisi
	procMoveToEx.Call(memDC, uintptr(centerX-size), uintptr(centerY-size), 0)
	procLineTo.Call(memDC, uintptr(centerX+size), uintptr(centerY+size))

	// Sol Alt -> Sağ Üst Çizgisi
	procMoveToEx.Call(memDC, uintptr(centerX-size), uintptr(centerY+size), 0)
	procLineTo.Call(memDC, uintptr(centerX+size), uintptr(centerY-size))

	// Kalem Temizliği
	procSelectObject.Call(memDC, oldPen)
	procDeleteObject.Call(pen)
}

// JSON Dosyasından Offset'leri Yükle
func loadOffsets(filePath string) error {
	data, err := os.ReadFile(filePath)
	if err != nil {
		return err
	}
	return json.Unmarshal(data, &offsets)
}

func main() {
	// 1. Offset'leri JSON'dan oku
	err := loadOffsets("offsets.json")
	if err != nil {
		fmt.Println("Offset dosyası okunamadı:", err)
		return
	}
	fmt.Println("[+] Offsetler yüklendi. m_pBulletServices:", offsets.M_pBulletServices, "m_totalHitsOnServer:", offsets.M_totalHitsOnServer)

	// Ekran Çözünürlüğünü Al
	screenWidthRes, _, _ := procGetSystemMetrics.Call(0)
	screenHeightRes, _, _ := procGetSystemMetrics.Call(1)
	screenWidth := int32(screenWidthRes)
	screenHeight := int32(screenHeightRes)

	/*
	   NOT: Burada kendi process açma ve overlay pencere oluşturma kodlarınız yer alır.
	   Örnek olarak ana güncelleme ve çizim döngüsü aşağıda sunulmuştur.
	*/

	// Ana Render / Memory Loop
	for {
		// LocalPlayerPawn Adresini Oku
		localPawn := ReadMemory[uintptr](clientDLL + offsets.DwLocalPlayerPawn)

		// 1. Hitmarker Mantığını Güncelle
		UpdateHitmarker(localPawn)

		// 2. Çizim İşlemleri (Örnek DC Yapısı)
		/*
			// memDC oluşturma ve double buffering çizimleri...
			DrawHitmarker(memDC, screenWidth, screenHeight)
			// BitBlt ile ekrana kopyalama...
		*/

		time.Sleep(1 * time.Millisecond)
	}
}
