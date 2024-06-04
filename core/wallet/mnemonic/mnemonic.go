// Package bip39 is the official Golang implementation of the BIP39 spec.
//
// The official BIP39 spec can be found at
// https://github.com/bitcoin/bips/blob/master/bip-0039.mediawiki
package mnemonic

import (
	"crypto/rand"
	"crypto/sha256"
	"crypto/sha512"
	"encoding/binary"
	"errors"
	"fmt"
	"math/big"
	"strings"

	"github.com/johngb/langreg"
	"golang.org/x/crypto/pbkdf2"

	"corepool/core/wallet/mnemonic/wordlists"
)

var (
	// Some bitwise operands for working with big.Ints
	last11BitsMask          = big.NewInt(2047)
	rightShift11BitsDivider = big.NewInt(2048)
	bigOne                  = big.NewInt(1)
	bigTwo                  = big.NewInt(2)

	// used to isolate the checksum bits from the entropy+checksum byte array
	wordLengthChecksumMasksMapping = map[int]*big.Int{
		12: big.NewInt(15),
		15: big.NewInt(31),
		18: big.NewInt(63),
		21: big.NewInt(127),
		24: big.NewInt(255),
	}
	// used to use only the desired x of 8 available checksum bits.
	// 256 bit (word length 24) requires all 8 bits of the checksum,
	// and thus no shifting is needed for it (we would get a divByZero crash if we did)
	wordLengthChecksumShiftMapping = map[int]*big.Int{
		12: big.NewInt(16),
		15: big.NewInt(8),
		18: big.NewInt(4),
		21: big.NewInt(2),
	}

	// wordList is the set of words to use
	wordList map[string][]string
)

var (
	// ErrInvalidMnemonic is returned when trying to use a malformed mnemonic.
	ErrInvalidMnemonic = errors.New("Invalid menomic")

	// ErrEntropyLengthInvalid is returned when trying to use an entropy set with
	// an invalid size.
	ErrEntropyLengthInvalid = errors.New("Entropy length must be [128, 256] and a multiple of 32")

	// ErrValidatedSeedLengthMismatch is returned when a validated seed is not the
	// same size as the given seed. This should never happen is present only as a
	// sanity assertion.
	ErrValidatedSeedLengthMismatch = errors.New("Seed length does not match validated seed length")

	// ErrChecksumIncorrect is returned when entropy has the incorrect checksum.
	ErrChecksumIncorrect = errors.New("Checksum incorrect")

	// ErrLanguageTypeIncorrect is return when find incorrect language type
	ErrLanguageTypeIncorrect = errors.New("Language Type Incorrect")

	// ErrLanguageTypeIncorrect is return when find incorrect language type
	ErrLanguageTypeUnsupported = errors.New("Language Type Unsupported")
)

func init() {
	wordList = map[string][]string{
		"zh_CN": wordlists.ChineseSimplified,
		"zh_TW": wordlists.ChineseTraditional,
		"en":    wordlists.English,
		"it":    wordlists.Italian,
		"ja":    wordlists.Japanese,
		"ko":    wordlists.Korean,
		"es":    wordlists.Spanish,
	}
}

// SetWordList sets the list of words to use for mnemonics. Currently the list
// that is set is used package-wide.
func SetWordMap(language string) (map[string]int, error) {
	if !isLanguageValid(language) {
		return nil, ErrLanguageTypeIncorrect
	}
	words, ok := wordList[language]
	if !ok {
		return nil, ErrLanguageTypeUnsupported
	}
	wordMap := map[string]int{}
	for i, v := range words {
		wordMap[v] = i
	}
	return wordMap, nil
}

// SetWordList sets the list of words to use for mnemonics. Currently the list
// that is set is used package-wide.
func SetWordList(language string) ([]string, error) {
	if !isLanguageValid(language) {
		return nil, ErrLanguageTypeIncorrect
	}
	words, ok := wordList[language]
	if !ok {
		return nil, ErrLanguageTypeUnsupported
	}

	return words, nil
}

// NewEntropy will create random entropy bytes
// so long as the requested size bitSize is an appropriate size.
//
// bitSize has to be a multiple 32 and be within the inclusive range of {128, 256}
func NewEntropy(bitSize int) ([]byte, error) {
	err := validateEntropyBitSize(bitSize)
	if err != nil {
		return nil, err
	}

	entropy := make([]byte, bitSize/8)
	_, err = rand.Read(entropy)
	return entropy, err
}

// EntropyFromMnemonic takes a mnemonic generated by this library,
// and returns the input entropy used to generate the given mnemonic.
// An error is returned if the given mnemonic is invalid.
func EntropyFromMnemonic(mnemonic string, language string) ([]byte, error) {
	mnemonicSlice, isValid := splitMnemonicWords(mnemonic)
	if !isValid {
		return nil, errors.New("Invalid mnemonic")
	}
	wordMap, err := SetWordMap(language)
	if err != nil {
		return nil, err
	}
	b := big.NewInt(0)
	for _, v := range mnemonicSlice {
		index, found := wordMap[v]
		if found == false {
			return nil, fmt.Errorf("word `%v` not found in reverse map", v)
		}
		var wordBytes [2]byte
		binary.BigEndian.PutUint16(wordBytes[:], uint16(index))
		b = b.Mul(b, rightShift11BitsDivider)
		b = b.Or(b, big.NewInt(0).SetBytes(wordBytes[:]))
	}

	checksum := big.NewInt(0)
	checksumMask := wordLengthChecksumMasksMapping[len(mnemonicSlice)]
	checksum = checksum.And(b, checksumMask)

	b.Div(b, big.NewInt(0).Add(checksumMask, bigOne))
	entropy := b.Bytes()
	// pad entropy if needed
	entropy = padByteSlice(entropy, len(mnemonicSlice)/3*4)

	// generate the checksum once again, mask and ensure it equals the checksum we got from the mneomnic
	entropyChecksumBytes := computeChecksum(entropy)
	entropyChecksum := big.NewInt(int64(entropyChecksumBytes[0]))
	if l := len(mnemonicSlice); l != 24 {
		checksumShift := wordLengthChecksumShiftMapping[l]
		entropyChecksum.Div(entropyChecksum, checksumShift)
	}

	if checksum.Cmp(entropyChecksum) != 0 {
		return nil, errors.New("mnemonic's entropy doesn't match its checksum")
	}

	// return (padded) entropy
	return entropy, nil
}

// NewMnemonic will return a string consisting of the mnemonic words for
// the given entropy.
// If the provide entropy is invalid, an error will be returned.
func NewMnemonic(entropy []byte, language string) (string, error) {
	wordList, err := SetWordList(language)
	if err != nil {
		return "", err
	}
	// Compute some lengths for convenience
	entropyBitLength := len(entropy) * 8
	checksumBitLength := entropyBitLength / 32
	sentenceLength := (entropyBitLength + checksumBitLength) / 11

	err = validateEntropyBitSize(entropyBitLength)
	if err != nil {
		return "", err
	}

	// Add checksum to entropy
	entropy = addChecksum(entropy)

	// Break entropy up into sentenceLength chunks of 11 bits
	// For each word AND mask the rightmost 11 bits and find the word at that index
	// Then bitshift entropy 11 bits right and repeat
	// Add to the last empty slot so we can work with LSBs instead of MSB

	// Entropy as an int so we can bitmask without worrying about bytes slices
	entropyInt := new(big.Int).SetBytes(entropy)

	// Slice to hold words in
	words := make([]string, sentenceLength)

	// Throw away big int for AND masking
	word := big.NewInt(0)

	for i := sentenceLength - 1; i >= 0; i-- {
		// Get 11 right most bits and bitshift 11 to the right for next time
		word.And(entropyInt, last11BitsMask)
		entropyInt.Div(entropyInt, rightShift11BitsDivider)

		// Get the bytes representing the 11 bits as a 2 byte slice
		wordBytes := padByteSlice(word.Bytes(), 2)

		// Convert bytes to an index and add that word to the list
		words[i] = wordList[binary.BigEndian.Uint16(wordBytes)]
	}

	return strings.Join(words, " "), nil
}

// MnemonicToByteArray takes a mnemonic string and turns it into a byte array
// suitable for creating another mnemonic.
// An error is returned if the mnemonic is invalid.
func MnemonicToByteArray(mnemonic string, language string, raw ...bool) ([]byte, error) {
	var (
		mnemonicSlice    = strings.Split(mnemonic, " ")
		entropyBitSize   = len(mnemonicSlice) * 11
		checksumBitSize  = entropyBitSize % 32
		fullByteSize     = (entropyBitSize-checksumBitSize)/8 + 1
		checksumByteSize = fullByteSize - (fullByteSize % 4)
	)
	wordMap, err := SetWordMap(language)
	if err != nil {
		return nil, err
	}
	// Pre validate that the mnemonic is well formed and only contains words that
	// are present in the word list
	if !IsMnemonicValid(mnemonic, language) {
		return nil, ErrInvalidMnemonic
	}

	// Convert word indices to a `big.Int` representing the entropy
	checksummedEntropy := big.NewInt(0)
	modulo := big.NewInt(2048)
	for _, v := range mnemonicSlice {
		index := big.NewInt(int64(wordMap[v]))
		checksummedEntropy.Mul(checksummedEntropy, modulo)
		checksummedEntropy.Add(checksummedEntropy, index)
	}

	// Calculate the unchecksummed entropy so we can validate that the checksum is
	// correct
	checksumModulo := big.NewInt(0).Exp(bigTwo, big.NewInt(int64(checksumBitSize)), nil)
	rawEntropy := big.NewInt(0).Div(checksummedEntropy, checksumModulo)

	// Convert `big.Int`s to byte padded byte slices
	rawEntropyBytes := padByteSlice(rawEntropy.Bytes(), checksumByteSize)
	checksummedEntropyBytes := padByteSlice(checksummedEntropy.Bytes(), fullByteSize)

	// Validate that the checksum is correct
	newChecksummedEntropyBytes := padByteSlice(addChecksum(rawEntropyBytes), fullByteSize)
	if !compareByteSlices(checksummedEntropyBytes, newChecksummedEntropyBytes) {
		return nil, ErrChecksumIncorrect
	}

	if raw == nil {
		return checksummedEntropyBytes, nil
	}
	if raw[0] == true {
		return rawEntropyBytes, nil
	}
	return checksummedEntropyBytes, nil
}

// NewSeedWithErrorChecking creates a hashed seed output given the mnemonic string and a password.
// An error is returned if the mnemonic is not convertible to a byte array.
func NewSeedWithErrorChecking(mnemonic string, password string, language string) ([]byte, error) {
	_, err := MnemonicToByteArray(mnemonic, language)
	if err != nil {
		return nil, err
	}
	return NewSeed(mnemonic, password), nil
}

// NewSeed creates a hashed seed output given a provided string and password.
// No checking is performed to validate that the string provided is a valid mnemonic.
func NewSeed(mnemonic string, password string) []byte {
	return pbkdf2.Key([]byte(mnemonic), []byte("mnemonic"+password), 2048, 64, sha512.New)
}

func isLanguageValid(language string) bool {
	if len(language) != 2 && len(language) != 5 {
		return false
	}
	if len(language) == 5 && !langreg.IsValidLangRegCode(language) {
		return false
	}
	if len(language) == 2 && !langreg.IsValidLanguageCode(language) {
		return false
	}
	return true
}

// IsMnemonicValid attempts to verify that the provided mnemonic is valid.
// Validity is determined by both the number of words being appropriate,
// and that all the words in the mnemonic are present in the word list.
func IsMnemonicValid(mnemonic string, language string) bool {
	// Create a list of all the words in the mnemonic sentence
	words := strings.Fields(mnemonic)

	// Get word count
	wordCount := len(words)

	// The number of words should be 12, 15, 18, 21 or 24
	if wordCount%3 != 0 || wordCount < 12 || wordCount > 24 {
		return false
	}
	wordMap, err := SetWordMap(language)
	if err != nil {
		return false
	}
	// Check if all words belong in the wordlist
	for _, word := range words {
		if _, ok := wordMap[word]; !ok {
			return false
		}
	}

	return true
}

// Appends to data the first (len(data) / 32)bits of the result of sha256(data)
// Currently only supports data up to 32 bytes
func addChecksum(data []byte) []byte {
	// Get first byte of sha256
	hash := computeChecksum(data)
	firstChecksumByte := hash[0]

	// len() is in bytes so we divide by 4
	checksumBitLength := uint(len(data) / 4)

	// For each bit of check sum we want we shift the data one the left
	// and then set the (new) right most bit equal to checksum bit at that index
	// staring from the left
	dataBigInt := new(big.Int).SetBytes(data)
	for i := uint(0); i < checksumBitLength; i++ {
		// Bitshift 1 left
		dataBigInt.Mul(dataBigInt, bigTwo)

		// Set rightmost bit if leftmost checksum bit is set
		if uint8(firstChecksumByte&(1<<(7-i))) > 0 {
			dataBigInt.Or(dataBigInt, bigOne)
		}
	}

	return dataBigInt.Bytes()
}

func computeChecksum(data []byte) []byte {
	hasher := sha256.New()
	hasher.Write(data)
	return hasher.Sum(nil)
}

// validateEntropyBitSize ensures that entropy is the correct size for being a
// mnemonic.
func validateEntropyBitSize(bitSize int) error {
	if (bitSize%32) != 0 || bitSize < 128 || bitSize > 256 {
		return ErrEntropyLengthInvalid
	}
	return nil
}

// padByteSlice returns a byte slice of the given size with contents of the
// given slice left padded and any empty spaces filled with 0's.
func padByteSlice(slice []byte, length int) []byte {
	offset := length - len(slice)
	if offset <= 0 {
		return slice
	}
	newSlice := make([]byte, length)
	copy(newSlice[offset:], slice)
	return newSlice
}

// compareByteSlices returns true of the byte slices have equal contents and
// returns false otherwise.
func compareByteSlices(a, b []byte) bool {
	if len(a) != len(b) {
		return false
	}
	for i := range a {
		if a[i] != b[i] {
			return false
		}
	}
	return true
}

func splitMnemonicWords(mnemonic string) ([]string, bool) {
	// Create a list of all the words in the mnemonic sentence
	words := strings.Fields(mnemonic)

	//Get num of words
	numOfWords := len(words)

	// The number of words should be 12, 15, 18, 21 or 24
	if numOfWords%3 != 0 || numOfWords < 12 || numOfWords > 24 {
		return nil, false
	}
	return words, true
}
