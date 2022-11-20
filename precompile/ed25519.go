package precompile

import (
	"crypto/ed25519"
	"encoding/json"
	"errors"
	"fmt"
	"math/big"
	"strings"

	"github.com/ava-labs/subnet-evm/accounts/abi"
	"github.com/ava-labs/subnet-evm/vmerrs"

	"github.com/ethereum/go-ethereum/common"
)

const (
	VerifyGasCost uint64 = 40_000

	// ED25519RawABI contains the raw ABI of ED25519 contract.
	ED25519RawABI = "[{\"inputs\":[{\"internalType\":\"bytes32\",\"name\":\"signature\",\"type\":\"bytes32\"},{\"internalType\":\"bytes32\",\"name\":\"message\",\"type\":\"bytes32\"},{\"internalType\":\"bytes32\",\"name\":\"publicKey\",\"type\":\"bytes32\"}],\"name\":\"verify\",\"outputs\":[{\"internalType\": \"bool\",\"name\": \"\",\"type\": \"bool\"}],\"stateMutability\":\"nonpayable\",\"type\":\"function\"}]"
)

// CUSTOM CODE STARTS HERE
// Reference imports to suppress errors from unused imports. This code and any unnecessary imports can be removed.
var (
	_ = errors.New
	_ = big.NewInt
	_ = strings.NewReader
	_ = fmt.Printf
)

// Singleton StatefulPrecompiledContract and signatures.
var (
	_ StatefulPrecompileConfig = &ED25519Config{}

	ED25519ABI abi.ABI // will be initialized by init function

	ED25519Precompile StatefulPrecompiledContract // will be initialized by init function
)

// ED25519Config implements the StatefulPrecompileConfig
// interface while adding in the ED25519 specific precompile address.
type ED25519Config struct {
	UpgradeableConfig
}

func init() {
	parsed, err := abi.JSON(strings.NewReader(ED25519RawABI))
	if err != nil {
		panic(err)
	}
	ED25519ABI = parsed

	ED25519Precompile = createED25519Precompile(ED25519Address)
}

// NewED25519Config returns a config for a network upgrade at [blockTimestamp] that enables
// ED25519 .
func NewED25519Config(blockTimestamp *big.Int) *ED25519Config {
	return &ED25519Config{

		UpgradeableConfig: UpgradeableConfig{BlockTimestamp: blockTimestamp},
	}
}

// NewDisableED25519Config returns config for a network upgrade at [blockTimestamp]
// that disables ED25519.
func NewDisableED25519Config(blockTimestamp *big.Int) *ED25519Config {
	return &ED25519Config{
		UpgradeableConfig: UpgradeableConfig{
			BlockTimestamp: blockTimestamp,
			Disable:        true,
		},
	}
}

// Equal returns true if [s] is a [*ED25519Config] and it has been configured identical to [c].
func (c *ED25519Config) Equal(s StatefulPrecompileConfig) bool {
	// typecast before comparison
	other, ok := (s).(*ED25519Config)
	if !ok {
		return false
	}
	// CUSTOM CODE STARTS HERE
	// modify this boolean accordingly with your custom ED25519Config, to check if [other] and the current [c] are equal
	// if ED25519Config contains only UpgradeableConfig  you can skip modifying it.
	equals := c.UpgradeableConfig.Equal(&other.UpgradeableConfig)
	return equals
}

// Address returns the address of the ED25519. Addresses reside under the precompile/params.go
// Select a non-conflicting address and set it in the params.go.
func (c *ED25519Config) Address() common.Address {
	return ED25519Address
}

// Configure configures [state] with the initial configuration.
func (c *ED25519Config) Configure(_ ChainConfig, state StateDB, _ BlockContext) {
	// CUSTOM CODE STARTS HERE
}

// String returns a string representation of the ED25519Config.
func (c *ED25519Config) String() string {
	bytes, _ := json.Marshal(c)
	return string(bytes)
}

// Contract returns the singleton stateful precompiled contract to be used for ED25519.
func (c *ED25519Config) Contract() StatefulPrecompiledContract {
	return ED25519Precompile
}

// Verify tries to verify ED25519Config and returns an error accordingly.
func (c *ED25519Config) Verify() error {

	// CUSTOM CODE STARTS HERE
	// Add your own custom verify code for ED25519Config here
	// and return an error accordingly
	return nil
}

// UnpackSetProtectionInput attempts to unpack [input] into the *big.Int type argument
// assumes that [input] does not include selector (omits first 4 func signature bytes)
func UnpackVerifyInput(input []byte) (*[]byte, *[]byte, *[]byte, error) {
	res, err := ED25519ABI.UnpackInput("verify", input)
	if err != nil {
		return nil, nil, nil, err
	}
	unpackedSignature := *abi.ConvertType(res[0], new(*[]byte)).(**[]byte)
	unpackedMessage := *abi.ConvertType(res[1], new(*[]byte)).(**[]byte)
	unpackedPublicKey := *abi.ConvertType(res[2], new(*[]byte)).(**[]byte)
	return unpackedSignature, unpackedMessage, unpackedPublicKey, nil
}

// PackSetProtection packs [protection] of type *big.Int into the appropriate arguments for setProtection.
// the packed bytes include selector (first 4 func signature bytes).
// This function is mostly used for tests.
func PackVerify(signature []byte, message []byte, publicKey []byte) ([]byte, error) {
	return ED25519ABI.Pack("verify", signature, message, publicKey)
}

func verify(accessibleState PrecompileAccessibleState, caller common.Address, addr common.Address, input []byte, suppliedGas uint64, readOnly bool) (ret []byte, remainingGas uint64, err error) {
	if remainingGas, err = deductGas(suppliedGas, VerifyGasCost); err != nil {
		return nil, 0, err
	}
	if readOnly {
		return nil, remainingGas, vmerrs.ErrWriteProtection
	}
	// attempts to unpack [input] into the arguments to the SetProtectionInput.
	// Assumes that [input] does not include selector
	// You can use unpacked [inputStruct] variable in your code
	signature, message, publicKey, err := UnpackVerifyInput(input)
	if err != nil {
		return nil, remainingGas, err
	}

	// cast publicKey to ed25519 PublicKey
	var pubKey ed25519.PublicKey = *publicKey

	// verify signature
	if !ed25519.Verify(pubKey, *message, *signature) {
		// return false
		return []byte{0}, remainingGas, nil
	}

	// Return true if signature is valid
	return []byte{1}, remainingGas, nil
}

// createED25519Precompile returns a StatefulPrecompiledContract with getters and setters for the precompile.

func createED25519Precompile(precompileAddr common.Address) StatefulPrecompiledContract {
	var functions []*statefulPrecompileFunction

	methodVerify, ok := ED25519ABI.Methods["verify"]
	if !ok {
		panic("given method does not exist in the ABI")
	}
	functions = append(functions, newStatefulPrecompileFunction(methodVerify.ID, verify))

	// Construct the contract with no fallback function.
	contract := newStatefulPrecompileWithFunctionSelectors(nil, functions)
	return contract
}
