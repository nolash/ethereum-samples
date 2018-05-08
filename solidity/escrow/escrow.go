// Code generated - DO NOT EDIT.
// This file is a generated binding and any manual changes will be lost.

package escrow

import (
	"math/big"
	"strings"

	"github.com/ethereum/go-ethereum/accounts/abi/bind"
	ethereum "github.com/ethereum/go-ethereum"
	"github.com/ethereum/go-ethereum/accounts/abi"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core/types"
	"github.com/ethereum/go-ethereum/event"
)

// EscrowABI is the input ABI used to generate the binding from.
const EscrowABI = "[{\"constant\":false,\"inputs\":[{\"name\":\"_id\",\"type\":\"uint256\"}],\"name\":\"participantPut\",\"outputs\":[],\"payable\":true,\"stateMutability\":\"payable\",\"type\":\"function\"},{\"constant\":false,\"inputs\":[{\"name\":\"_id\",\"type\":\"uint256\"}],\"name\":\"participantCancel\",\"outputs\":[],\"payable\":false,\"stateMutability\":\"nonpayable\",\"type\":\"function\"},{\"constant\":false,\"inputs\":[{\"name\":\"_id\",\"type\":\"uint256\"}],\"name\":\"participantPayout\",\"outputs\":[{\"name\":\"\",\"type\":\"bool\"}],\"payable\":false,\"stateMutability\":\"nonpayable\",\"type\":\"function\"},{\"constant\":false,\"inputs\":[{\"name\":\"_id\",\"type\":\"uint256\"}],\"name\":\"participantRelease\",\"outputs\":[],\"payable\":false,\"stateMutability\":\"nonpayable\",\"type\":\"function\"},{\"constant\":false,\"inputs\":[{\"name\":\"_name\",\"type\":\"bytes8\"}],\"name\":\"createEscrow\",\"outputs\":[],\"payable\":false,\"stateMutability\":\"nonpayable\",\"type\":\"function\"},{\"constant\":true,\"inputs\":[],\"name\":\"seq\",\"outputs\":[{\"name\":\"\",\"type\":\"uint256\"}],\"payable\":false,\"stateMutability\":\"view\",\"type\":\"function\"},{\"constant\":false,\"inputs\":[{\"name\":\"_seq\",\"type\":\"uint256\"},{\"name\":\"_amount\",\"type\":\"uint80\"},{\"name\":\"_who\",\"type\":\"address\"},{\"name\":\"_beneficiary\",\"type\":\"address\"}],\"name\":\"addParticipant\",\"outputs\":[],\"payable\":false,\"stateMutability\":\"nonpayable\",\"type\":\"function\"},{\"anonymous\":false,\"inputs\":[{\"indexed\":false,\"name\":\"seq\",\"type\":\"uint256\"},{\"indexed\":false,\"name\":\"_name\",\"type\":\"bytes8\"}],\"name\":\"Created\",\"type\":\"event\"},{\"anonymous\":false,\"inputs\":[{\"indexed\":false,\"name\":\"seq\",\"type\":\"uint256\"},{\"indexed\":false,\"name\":\"participant\",\"type\":\"address\"}],\"name\":\"ParticipantPaid\",\"type\":\"event\"},{\"anonymous\":false,\"inputs\":[{\"indexed\":false,\"name\":\"seq\",\"type\":\"uint256\"},{\"indexed\":false,\"name\":\"participant\",\"type\":\"address\"}],\"name\":\"ParticipantReleased\",\"type\":\"event\"},{\"anonymous\":false,\"inputs\":[{\"indexed\":false,\"name\":\"seq\",\"type\":\"uint256\"},{\"indexed\":false,\"name\":\"participant\",\"type\":\"address\"}],\"name\":\"ParticipantCancelled\",\"type\":\"event\"},{\"anonymous\":false,\"inputs\":[{\"indexed\":false,\"name\":\"seq\",\"type\":\"uint256\"}],\"name\":\"Finished\",\"type\":\"event\"}]"

// EscrowBin is the compiled bytecode used for deploying new contracts.
const EscrowBin = `0x608060405234801561001057600080fd5b50610f02806100206000396000f3006080604052600436106100825763ffffffff7c0100000000000000000000000000000000000000000000000000000000600035041663136b27c78114610087578063204e71b914610094578063422ee44a146100ac57806343ae9547146100d857806344413984146100f05780636857ab401461012357806399757f9b1461014a575b600080fd5b610092600435610183565b005b3480156100a057600080fd5b5061009260043561031d565b3480156100b857600080fd5b506100c4600435610502565b604080519115158252519081900360200190f35b3480156100e457600080fd5b50610092600435610999565b3480156100fc57600080fd5b5061009277ffffffffffffffffffffffffffffffffffffffffffffffff1960043516610a7e565b34801561012f57600080fd5b50610138610bfa565b60408051918252519081900360200190f35b34801561015657600080fd5b5061009260043569ffffffffffffffffffff60243516600160a060020a0360443581169060643516610c00565b60028054600091908390811061019557fe5b6000918252602080832033600160a060020a0390811685526003939093020160010190526040909120546b01000000000000000000000090041614156101da57600080fd5b346002828154811015156101ea57fe5b60009182526020808320600160a060020a03331684526001600390930201919091019052604090205469ffffffffffffffffffff16111561022a57600080fd5b3460028281548110151561023a57fe5b6000918252602080832033600160a060020a03168452600392909202909101600101905260409020805469ffffffffffffffffffff191669ffffffffffffffffffff929092169190911790556002805481908390811061029657fe5b6000918252602080832033600160a060020a031680855260039390930201600101815260409283902080546aff000000000000000000001916605060020a60ff96909616959095029490941790935581518481529283015280517fc7d538e920165422bcfcc5aded5d1866e4ddca05191db88a2cf99b0377dcecee9281900390910190a150565b600080600460028481548110151561033157fe5b6000918252602080832033600160a060020a0316845260039290920290910160010190526040902054605060020a900460ff161061036e57600080fd5b5060005b600280548490811061038057fe5b9060005260206000209060030201600201805490508160ff1610156104375760028054849081106103ad57fe5b906000526020600020906003020160010160006002858154811015156103cf57fe5b90600052602060002090600302016002018360ff168154811015156103f057fe5b600091825260208083209190910154600160a060020a03168352820192909252604001902054605060020a900460ff166003141561042f576001909101905b600101610372565b600280548490811061044557fe5b600091825260209091206002600390920201015460ff8316141561046857600080fd5b600560028481548110151561047957fe5b6000918252602080832033600160a060020a031680855260039390930201600101815260409283902080546aff000000000000000000001916605060020a60ff96909616959095029490941790935581518681529283015280517f91802c983b6a8e1758dd1f19589461b58e3a7c6209dbb1b45c4f382f052a1d529281900390910190a1505050565b60008060008060006002808781548110151561051a57fe5b6000918252602080832033600160a060020a0316845260039290920290910160010190526040902054605060020a900460ff161161055757600080fd5b60039350600092505b600280548790811061056e57fe5b9060005260206000209060030201600201805490508360ff1610156106b757600280548790811061059b57fe5b906000526020600020906003020160010160006002888154811015156105bd57fe5b90600052602060002090600302016002018560ff168154811015156105de57fe5b600091825260208083209190910154600160a060020a03168352820192909252604001902054605060020a900460ff166005141561061f57600593506106b7565b600280548790811061062d57fe5b9060005260206000209060030201600101600060028881548110151561064f57fe5b90600052602060002090600302016002018560ff1681548110151561067057fe5b600091825260208083209190910154600160a060020a03168352820192909252604001902054605060020a900460ff166003146106ac57600080fd5b600190920191610560565b60028054879081106106c557fe5b60009182526020808320600160a060020a0333168452600160039093020191909101905260409020546002805469ffffffffffffffffffff9092169350908790811061070d57fe5b6000918252602080832033600160a060020a0316845260039290920290910160010190526040902054605060020a900460ff166005141561074f5750336107ab565b8360ff166005141561076057600080fd5b600280548790811061076e57fe5b6000918252602080832033600160a060020a0390811685526003939093020160010190526040909120546b01000000000000000000000090041690505b60046002878154811015156107bc57fe5b60009182526020808320600160a060020a0333168452600160039093020191909101905260408120805460ff93909316605060020a026aff000000000000000000001990931692909217909155600280548890811061081757fe5b6000918252602080832033600160a060020a03168085526003939093020160010190526040909120805469ffffffffffffffffffff191669ffffffffffffffffffff9390931692909217909155600280546108fc91908990811061087757fe5b60009182526020808320600160a060020a03331684526001600390930201919091019052604080822054905169ffffffffffffffffffff90911680159093029291818181858888f19350505050151561098b57816002878154811015156108da57fe5b6000918252602080832033600160a060020a03168452600392909202909101600101905260409020805469ffffffffffffffffffff191669ffffffffffffffffffff92909216919091179055600280548591908890811061093757fe5b60009182526020808320600160a060020a0333168452600160039093020191909101905260408120805460ff93909316605060020a026aff0000000000000000000019909316929092179091559450610990565b600194505b50505050919050565b60028054829081106109a757fe5b6000918252602080832033600160a060020a0316845260039290920290910160010190526040902054605060020a900460ff166002146109e657600080fd5b60036002828154811015156109f757fe5b6000918252602080832033600160a060020a031680855260039390930201600101815260409283902080546aff000000000000000000001916605060020a60ff96909616959095029490941790935581518481529283015280517f6f29a1913da999f186768f08d8f2ed16f36182d1e2073a748f825b38a21f79909281900390910190a150565b610a86610df1565b6000805460019081018255600160a060020a03338116602080860191825277ffffffffffffffffffffffffffffffffffffffffffffffff1987168652600280549485018082559552855160039094027f405787fa12a823e0f2b7631cc41b3ba8828b3321ca811111fa75cd3aa3bb5ace81018054935190941668010000000000000000027fffffffff0000000000000000000000000000000000000000ffffffffffffffff780100000000000000000000000000000000000000000000000090960467ffffffffffffffff199094169390931794909416919091178255604085015180518694610b9c937f405787fa12a823e0f2b7631cc41b3ba8828b3321ca811111fa75cd3aa3bb5ad0909101920190610e10565b50506000546040805191825277ffffffffffffffffffffffffffffffffffffffffffffffff198616602083015280517f32e343b6c3095e099c5dd22ddf8c3fa6a5c5514a4286f78a7e8886374a33d4fe945091829003019150a15050565b60005481565b610c08610e82565b600280546000198701908110610c1a57fe5b600091825260209091206003909102015433600160a060020a03908116680100000000000000009092041614610c4f57600080fd5b600280546000198701908110610c6157fe5b60009182526020808320600160a060020a038716845260039290920290910160010190526040902054605060020a900460ff1615610c9e57600080fd5b69ffffffffffffffffffff84168152600160a060020a038216604082015260016020820152600280548291906000198801908110610cd857fe5b60009182526020808320600160a060020a03808916855260016003909402909101929092018152604092839020845181549286015195909401519092166b010000000000000000000000027fff0000000000000000000000000000000000000000ffffffffffffffffffffff60ff909516605060020a026aff000000000000000000001969ffffffffffffffffffff90951669ffffffffffffffffffff1990931692909217939093161792909216179055600280546000198701908110610d9b57fe5b600091825260208083206003929092029091016002018054600181018255908352912001805473ffffffffffffffffffffffffffffffffffffffff1916600160a060020a03949094169390931790925550505050565b6040805160608181018352600080835260208301529181019190915290565b828054828255906000526020600020908101928215610e72579160200282015b82811115610e72578251825473ffffffffffffffffffffffffffffffffffffffff1916600160a060020a03909116178255602090920191600190910190610e30565b50610e7e929150610ea2565b5090565b604080516060810182526000808252602082018190529181019190915290565b610ed391905b80821115610e7e57805473ffffffffffffffffffffffffffffffffffffffff19168155600101610ea8565b905600a165627a7a72305820f8fbbb338fc742e6dde6afd59ba982a9c1f0903763b2c6297f20eb51029fb54b0029`

// DeployEscrow deploys a new Ethereum contract, binding an instance of Escrow to it.
func DeployEscrow(auth *bind.TransactOpts, backend bind.ContractBackend) (common.Address, *types.Transaction, *Escrow, error) {
	parsed, err := abi.JSON(strings.NewReader(EscrowABI))
	if err != nil {
		return common.Address{}, nil, nil, err
	}
	address, tx, contract, err := bind.DeployContract(auth, parsed, common.FromHex(EscrowBin), backend)
	if err != nil {
		return common.Address{}, nil, nil, err
	}
	return address, tx, &Escrow{EscrowCaller: EscrowCaller{contract: contract}, EscrowTransactor: EscrowTransactor{contract: contract}, EscrowFilterer: EscrowFilterer{contract: contract}}, nil
}

// Escrow is an auto generated Go binding around an Ethereum contract.
type Escrow struct {
	EscrowCaller     // Read-only binding to the contract
	EscrowTransactor // Write-only binding to the contract
	EscrowFilterer   // Log filterer for contract events
}

// EscrowCaller is an auto generated read-only Go binding around an Ethereum contract.
type EscrowCaller struct {
	contract *bind.BoundContract // Generic contract wrapper for the low level calls
}

// EscrowTransactor is an auto generated write-only Go binding around an Ethereum contract.
type EscrowTransactor struct {
	contract *bind.BoundContract // Generic contract wrapper for the low level calls
}

// EscrowFilterer is an auto generated log filtering Go binding around an Ethereum contract events.
type EscrowFilterer struct {
	contract *bind.BoundContract // Generic contract wrapper for the low level calls
}

// EscrowSession is an auto generated Go binding around an Ethereum contract,
// with pre-set call and transact options.
type EscrowSession struct {
	Contract     *Escrow           // Generic contract binding to set the session for
	CallOpts     bind.CallOpts     // Call options to use throughout this session
	TransactOpts bind.TransactOpts // Transaction auth options to use throughout this session
}

// EscrowCallerSession is an auto generated read-only Go binding around an Ethereum contract,
// with pre-set call options.
type EscrowCallerSession struct {
	Contract *EscrowCaller // Generic contract caller binding to set the session for
	CallOpts bind.CallOpts // Call options to use throughout this session
}

// EscrowTransactorSession is an auto generated write-only Go binding around an Ethereum contract,
// with pre-set transact options.
type EscrowTransactorSession struct {
	Contract     *EscrowTransactor // Generic contract transactor binding to set the session for
	TransactOpts bind.TransactOpts // Transaction auth options to use throughout this session
}

// EscrowRaw is an auto generated low-level Go binding around an Ethereum contract.
type EscrowRaw struct {
	Contract *Escrow // Generic contract binding to access the raw methods on
}

// EscrowCallerRaw is an auto generated low-level read-only Go binding around an Ethereum contract.
type EscrowCallerRaw struct {
	Contract *EscrowCaller // Generic read-only contract binding to access the raw methods on
}

// EscrowTransactorRaw is an auto generated low-level write-only Go binding around an Ethereum contract.
type EscrowTransactorRaw struct {
	Contract *EscrowTransactor // Generic write-only contract binding to access the raw methods on
}

// NewEscrow creates a new instance of Escrow, bound to a specific deployed contract.
func NewEscrow(address common.Address, backend bind.ContractBackend) (*Escrow, error) {
	contract, err := bindEscrow(address, backend, backend, backend)
	if err != nil {
		return nil, err
	}
	return &Escrow{EscrowCaller: EscrowCaller{contract: contract}, EscrowTransactor: EscrowTransactor{contract: contract}, EscrowFilterer: EscrowFilterer{contract: contract}}, nil
}

// NewEscrowCaller creates a new read-only instance of Escrow, bound to a specific deployed contract.
func NewEscrowCaller(address common.Address, caller bind.ContractCaller) (*EscrowCaller, error) {
	contract, err := bindEscrow(address, caller, nil, nil)
	if err != nil {
		return nil, err
	}
	return &EscrowCaller{contract: contract}, nil
}

// NewEscrowTransactor creates a new write-only instance of Escrow, bound to a specific deployed contract.
func NewEscrowTransactor(address common.Address, transactor bind.ContractTransactor) (*EscrowTransactor, error) {
	contract, err := bindEscrow(address, nil, transactor, nil)
	if err != nil {
		return nil, err
	}
	return &EscrowTransactor{contract: contract}, nil
}

// NewEscrowFilterer creates a new log filterer instance of Escrow, bound to a specific deployed contract.
func NewEscrowFilterer(address common.Address, filterer bind.ContractFilterer) (*EscrowFilterer, error) {
	contract, err := bindEscrow(address, nil, nil, filterer)
	if err != nil {
		return nil, err
	}
	return &EscrowFilterer{contract: contract}, nil
}

// bindEscrow binds a generic wrapper to an already deployed contract.
func bindEscrow(address common.Address, caller bind.ContractCaller, transactor bind.ContractTransactor, filterer bind.ContractFilterer) (*bind.BoundContract, error) {
	parsed, err := abi.JSON(strings.NewReader(EscrowABI))
	if err != nil {
		return nil, err
	}
	return bind.NewBoundContract(address, parsed, caller, transactor, filterer), nil
}

// Call invokes the (constant) contract method with params as input values and
// sets the output to result. The result type might be a single field for simple
// returns, a slice of interfaces for anonymous returns and a struct for named
// returns.
func (_Escrow *EscrowRaw) Call(opts *bind.CallOpts, result interface{}, method string, params ...interface{}) error {
	return _Escrow.Contract.EscrowCaller.contract.Call(opts, result, method, params...)
}

// Transfer initiates a plain transaction to move funds to the contract, calling
// its default method if one is available.
func (_Escrow *EscrowRaw) Transfer(opts *bind.TransactOpts) (*types.Transaction, error) {
	return _Escrow.Contract.EscrowTransactor.contract.Transfer(opts)
}

// Transact invokes the (paid) contract method with params as input values.
func (_Escrow *EscrowRaw) Transact(opts *bind.TransactOpts, method string, params ...interface{}) (*types.Transaction, error) {
	return _Escrow.Contract.EscrowTransactor.contract.Transact(opts, method, params...)
}

// Call invokes the (constant) contract method with params as input values and
// sets the output to result. The result type might be a single field for simple
// returns, a slice of interfaces for anonymous returns and a struct for named
// returns.
func (_Escrow *EscrowCallerRaw) Call(opts *bind.CallOpts, result interface{}, method string, params ...interface{}) error {
	return _Escrow.Contract.contract.Call(opts, result, method, params...)
}

// Transfer initiates a plain transaction to move funds to the contract, calling
// its default method if one is available.
func (_Escrow *EscrowTransactorRaw) Transfer(opts *bind.TransactOpts) (*types.Transaction, error) {
	return _Escrow.Contract.contract.Transfer(opts)
}

// Transact invokes the (paid) contract method with params as input values.
func (_Escrow *EscrowTransactorRaw) Transact(opts *bind.TransactOpts, method string, params ...interface{}) (*types.Transaction, error) {
	return _Escrow.Contract.contract.Transact(opts, method, params...)
}

// Seq is a free data retrieval call binding the contract method 0x6857ab40.
//
// Solidity: function seq() constant returns(uint256)
func (_Escrow *EscrowCaller) Seq(opts *bind.CallOpts) (*big.Int, error) {
	var (
		ret0 = new(*big.Int)
	)
	out := ret0
	err := _Escrow.contract.Call(opts, out, "seq")
	return *ret0, err
}

// Seq is a free data retrieval call binding the contract method 0x6857ab40.
//
// Solidity: function seq() constant returns(uint256)
func (_Escrow *EscrowSession) Seq() (*big.Int, error) {
	return _Escrow.Contract.Seq(&_Escrow.CallOpts)
}

// Seq is a free data retrieval call binding the contract method 0x6857ab40.
//
// Solidity: function seq() constant returns(uint256)
func (_Escrow *EscrowCallerSession) Seq() (*big.Int, error) {
	return _Escrow.Contract.Seq(&_Escrow.CallOpts)
}

// AddParticipant is a paid mutator transaction binding the contract method 0x99757f9b.
//
// Solidity: function addParticipant(_seq uint256, _amount uint80, _who address, _beneficiary address) returns()
func (_Escrow *EscrowTransactor) AddParticipant(opts *bind.TransactOpts, _seq *big.Int, _amount *big.Int, _who common.Address, _beneficiary common.Address) (*types.Transaction, error) {
	return _Escrow.contract.Transact(opts, "addParticipant", _seq, _amount, _who, _beneficiary)
}

// AddParticipant is a paid mutator transaction binding the contract method 0x99757f9b.
//
// Solidity: function addParticipant(_seq uint256, _amount uint80, _who address, _beneficiary address) returns()
func (_Escrow *EscrowSession) AddParticipant(_seq *big.Int, _amount *big.Int, _who common.Address, _beneficiary common.Address) (*types.Transaction, error) {
	return _Escrow.Contract.AddParticipant(&_Escrow.TransactOpts, _seq, _amount, _who, _beneficiary)
}

// AddParticipant is a paid mutator transaction binding the contract method 0x99757f9b.
//
// Solidity: function addParticipant(_seq uint256, _amount uint80, _who address, _beneficiary address) returns()
func (_Escrow *EscrowTransactorSession) AddParticipant(_seq *big.Int, _amount *big.Int, _who common.Address, _beneficiary common.Address) (*types.Transaction, error) {
	return _Escrow.Contract.AddParticipant(&_Escrow.TransactOpts, _seq, _amount, _who, _beneficiary)
}

// CreateEscrow is a paid mutator transaction binding the contract method 0x44413984.
//
// Solidity: function createEscrow(_name bytes8) returns()
func (_Escrow *EscrowTransactor) CreateEscrow(opts *bind.TransactOpts, _name [8]byte) (*types.Transaction, error) {
	return _Escrow.contract.Transact(opts, "createEscrow", _name)
}

// CreateEscrow is a paid mutator transaction binding the contract method 0x44413984.
//
// Solidity: function createEscrow(_name bytes8) returns()
func (_Escrow *EscrowSession) CreateEscrow(_name [8]byte) (*types.Transaction, error) {
	return _Escrow.Contract.CreateEscrow(&_Escrow.TransactOpts, _name)
}

// CreateEscrow is a paid mutator transaction binding the contract method 0x44413984.
//
// Solidity: function createEscrow(_name bytes8) returns()
func (_Escrow *EscrowTransactorSession) CreateEscrow(_name [8]byte) (*types.Transaction, error) {
	return _Escrow.Contract.CreateEscrow(&_Escrow.TransactOpts, _name)
}

// ParticipantCancel is a paid mutator transaction binding the contract method 0x204e71b9.
//
// Solidity: function participantCancel(_id uint256) returns()
func (_Escrow *EscrowTransactor) ParticipantCancel(opts *bind.TransactOpts, _id *big.Int) (*types.Transaction, error) {
	return _Escrow.contract.Transact(opts, "participantCancel", _id)
}

// ParticipantCancel is a paid mutator transaction binding the contract method 0x204e71b9.
//
// Solidity: function participantCancel(_id uint256) returns()
func (_Escrow *EscrowSession) ParticipantCancel(_id *big.Int) (*types.Transaction, error) {
	return _Escrow.Contract.ParticipantCancel(&_Escrow.TransactOpts, _id)
}

// ParticipantCancel is a paid mutator transaction binding the contract method 0x204e71b9.
//
// Solidity: function participantCancel(_id uint256) returns()
func (_Escrow *EscrowTransactorSession) ParticipantCancel(_id *big.Int) (*types.Transaction, error) {
	return _Escrow.Contract.ParticipantCancel(&_Escrow.TransactOpts, _id)
}

// ParticipantPayout is a paid mutator transaction binding the contract method 0x422ee44a.
//
// Solidity: function participantPayout(_id uint256) returns(bool)
func (_Escrow *EscrowTransactor) ParticipantPayout(opts *bind.TransactOpts, _id *big.Int) (*types.Transaction, error) {
	return _Escrow.contract.Transact(opts, "participantPayout", _id)
}

// ParticipantPayout is a paid mutator transaction binding the contract method 0x422ee44a.
//
// Solidity: function participantPayout(_id uint256) returns(bool)
func (_Escrow *EscrowSession) ParticipantPayout(_id *big.Int) (*types.Transaction, error) {
	return _Escrow.Contract.ParticipantPayout(&_Escrow.TransactOpts, _id)
}

// ParticipantPayout is a paid mutator transaction binding the contract method 0x422ee44a.
//
// Solidity: function participantPayout(_id uint256) returns(bool)
func (_Escrow *EscrowTransactorSession) ParticipantPayout(_id *big.Int) (*types.Transaction, error) {
	return _Escrow.Contract.ParticipantPayout(&_Escrow.TransactOpts, _id)
}

// ParticipantPut is a paid mutator transaction binding the contract method 0x136b27c7.
//
// Solidity: function participantPut(_id uint256) returns()
func (_Escrow *EscrowTransactor) ParticipantPut(opts *bind.TransactOpts, _id *big.Int) (*types.Transaction, error) {
	return _Escrow.contract.Transact(opts, "participantPut", _id)
}

// ParticipantPut is a paid mutator transaction binding the contract method 0x136b27c7.
//
// Solidity: function participantPut(_id uint256) returns()
func (_Escrow *EscrowSession) ParticipantPut(_id *big.Int) (*types.Transaction, error) {
	return _Escrow.Contract.ParticipantPut(&_Escrow.TransactOpts, _id)
}

// ParticipantPut is a paid mutator transaction binding the contract method 0x136b27c7.
//
// Solidity: function participantPut(_id uint256) returns()
func (_Escrow *EscrowTransactorSession) ParticipantPut(_id *big.Int) (*types.Transaction, error) {
	return _Escrow.Contract.ParticipantPut(&_Escrow.TransactOpts, _id)
}

// ParticipantRelease is a paid mutator transaction binding the contract method 0x43ae9547.
//
// Solidity: function participantRelease(_id uint256) returns()
func (_Escrow *EscrowTransactor) ParticipantRelease(opts *bind.TransactOpts, _id *big.Int) (*types.Transaction, error) {
	return _Escrow.contract.Transact(opts, "participantRelease", _id)
}

// ParticipantRelease is a paid mutator transaction binding the contract method 0x43ae9547.
//
// Solidity: function participantRelease(_id uint256) returns()
func (_Escrow *EscrowSession) ParticipantRelease(_id *big.Int) (*types.Transaction, error) {
	return _Escrow.Contract.ParticipantRelease(&_Escrow.TransactOpts, _id)
}

// ParticipantRelease is a paid mutator transaction binding the contract method 0x43ae9547.
//
// Solidity: function participantRelease(_id uint256) returns()
func (_Escrow *EscrowTransactorSession) ParticipantRelease(_id *big.Int) (*types.Transaction, error) {
	return _Escrow.Contract.ParticipantRelease(&_Escrow.TransactOpts, _id)
}

// EscrowCreatedIterator is returned from FilterCreated and is used to iterate over the raw logs and unpacked data for Created events raised by the Escrow contract.
type EscrowCreatedIterator struct {
	Event *EscrowCreated // Event containing the contract specifics and raw log

	contract *bind.BoundContract // Generic contract to use for unpacking event data
	event    string              // Event name to use for unpacking event data

	logs chan types.Log        // Log channel receiving the found contract events
	sub  ethereum.Subscription // Subscription for errors, completion and termination
	done bool                  // Whether the subscription completed delivering logs
	fail error                 // Occurred error to stop iteration
}

// Next advances the iterator to the subsequent event, returning whether there
// are any more events found. In case of a retrieval or parsing error, false is
// returned and Error() can be queried for the exact failure.
func (it *EscrowCreatedIterator) Next() bool {
	// If the iterator failed, stop iterating
	if it.fail != nil {
		return false
	}
	// If the iterator completed, deliver directly whatever's available
	if it.done {
		select {
		case log := <-it.logs:
			it.Event = new(EscrowCreated)
			if err := it.contract.UnpackLog(it.Event, it.event, log); err != nil {
				it.fail = err
				return false
			}
			it.Event.Raw = log
			return true

		default:
			return false
		}
	}
	// Iterator still in progress, wait for either a data or an error event
	select {
	case log := <-it.logs:
		it.Event = new(EscrowCreated)
		if err := it.contract.UnpackLog(it.Event, it.event, log); err != nil {
			it.fail = err
			return false
		}
		it.Event.Raw = log
		return true

	case err := <-it.sub.Err():
		it.done = true
		it.fail = err
		return it.Next()
	}
}

// Error returns any retrieval or parsing error occurred during filtering.
func (it *EscrowCreatedIterator) Error() error {
	return it.fail
}

// Close terminates the iteration process, releasing any pending underlying
// resources.
func (it *EscrowCreatedIterator) Close() error {
	it.sub.Unsubscribe()
	return nil
}

// EscrowCreated represents a Created event raised by the Escrow contract.
type EscrowCreated struct {
	Seq  *big.Int
	Name [8]byte
	Raw  types.Log // Blockchain specific contextual infos
}

// FilterCreated is a free log retrieval operation binding the contract event 0x32e343b6c3095e099c5dd22ddf8c3fa6a5c5514a4286f78a7e8886374a33d4fe.
//
// Solidity: event Created(seq uint256, _name bytes8)
func (_Escrow *EscrowFilterer) FilterCreated(opts *bind.FilterOpts) (*EscrowCreatedIterator, error) {

	logs, sub, err := _Escrow.contract.FilterLogs(opts, "Created")
	if err != nil {
		return nil, err
	}
	return &EscrowCreatedIterator{contract: _Escrow.contract, event: "Created", logs: logs, sub: sub}, nil
}

// WatchCreated is a free log subscription operation binding the contract event 0x32e343b6c3095e099c5dd22ddf8c3fa6a5c5514a4286f78a7e8886374a33d4fe.
//
// Solidity: event Created(seq uint256, _name bytes8)
func (_Escrow *EscrowFilterer) WatchCreated(opts *bind.WatchOpts, sink chan<- *EscrowCreated) (event.Subscription, error) {

	logs, sub, err := _Escrow.contract.WatchLogs(opts, "Created")
	if err != nil {
		return nil, err
	}
	return event.NewSubscription(func(quit <-chan struct{}) error {
		defer sub.Unsubscribe()
		for {
			select {
			case log := <-logs:
				// New log arrived, parse the event and forward to the user
				event := new(EscrowCreated)
				if err := _Escrow.contract.UnpackLog(event, "Created", log); err != nil {
					return err
				}
				event.Raw = log

				select {
				case sink <- event:
				case err := <-sub.Err():
					return err
				case <-quit:
					return nil
				}
			case err := <-sub.Err():
				return err
			case <-quit:
				return nil
			}
		}
	}), nil
}

// EscrowFinishedIterator is returned from FilterFinished and is used to iterate over the raw logs and unpacked data for Finished events raised by the Escrow contract.
type EscrowFinishedIterator struct {
	Event *EscrowFinished // Event containing the contract specifics and raw log

	contract *bind.BoundContract // Generic contract to use for unpacking event data
	event    string              // Event name to use for unpacking event data

	logs chan types.Log        // Log channel receiving the found contract events
	sub  ethereum.Subscription // Subscription for errors, completion and termination
	done bool                  // Whether the subscription completed delivering logs
	fail error                 // Occurred error to stop iteration
}

// Next advances the iterator to the subsequent event, returning whether there
// are any more events found. In case of a retrieval or parsing error, false is
// returned and Error() can be queried for the exact failure.
func (it *EscrowFinishedIterator) Next() bool {
	// If the iterator failed, stop iterating
	if it.fail != nil {
		return false
	}
	// If the iterator completed, deliver directly whatever's available
	if it.done {
		select {
		case log := <-it.logs:
			it.Event = new(EscrowFinished)
			if err := it.contract.UnpackLog(it.Event, it.event, log); err != nil {
				it.fail = err
				return false
			}
			it.Event.Raw = log
			return true

		default:
			return false
		}
	}
	// Iterator still in progress, wait for either a data or an error event
	select {
	case log := <-it.logs:
		it.Event = new(EscrowFinished)
		if err := it.contract.UnpackLog(it.Event, it.event, log); err != nil {
			it.fail = err
			return false
		}
		it.Event.Raw = log
		return true

	case err := <-it.sub.Err():
		it.done = true
		it.fail = err
		return it.Next()
	}
}

// Error returns any retrieval or parsing error occurred during filtering.
func (it *EscrowFinishedIterator) Error() error {
	return it.fail
}

// Close terminates the iteration process, releasing any pending underlying
// resources.
func (it *EscrowFinishedIterator) Close() error {
	it.sub.Unsubscribe()
	return nil
}

// EscrowFinished represents a Finished event raised by the Escrow contract.
type EscrowFinished struct {
	Seq *big.Int
	Raw types.Log // Blockchain specific contextual infos
}

// FilterFinished is a free log retrieval operation binding the contract event 0x86954ecc0ae072157fcf7f87a425a1461295a4cc9cc3122d2efc73bf32d98e1a.
//
// Solidity: event Finished(seq uint256)
func (_Escrow *EscrowFilterer) FilterFinished(opts *bind.FilterOpts) (*EscrowFinishedIterator, error) {

	logs, sub, err := _Escrow.contract.FilterLogs(opts, "Finished")
	if err != nil {
		return nil, err
	}
	return &EscrowFinishedIterator{contract: _Escrow.contract, event: "Finished", logs: logs, sub: sub}, nil
}

// WatchFinished is a free log subscription operation binding the contract event 0x86954ecc0ae072157fcf7f87a425a1461295a4cc9cc3122d2efc73bf32d98e1a.
//
// Solidity: event Finished(seq uint256)
func (_Escrow *EscrowFilterer) WatchFinished(opts *bind.WatchOpts, sink chan<- *EscrowFinished) (event.Subscription, error) {

	logs, sub, err := _Escrow.contract.WatchLogs(opts, "Finished")
	if err != nil {
		return nil, err
	}
	return event.NewSubscription(func(quit <-chan struct{}) error {
		defer sub.Unsubscribe()
		for {
			select {
			case log := <-logs:
				// New log arrived, parse the event and forward to the user
				event := new(EscrowFinished)
				if err := _Escrow.contract.UnpackLog(event, "Finished", log); err != nil {
					return err
				}
				event.Raw = log

				select {
				case sink <- event:
				case err := <-sub.Err():
					return err
				case <-quit:
					return nil
				}
			case err := <-sub.Err():
				return err
			case <-quit:
				return nil
			}
		}
	}), nil
}

// EscrowParticipantCancelledIterator is returned from FilterParticipantCancelled and is used to iterate over the raw logs and unpacked data for ParticipantCancelled events raised by the Escrow contract.
type EscrowParticipantCancelledIterator struct {
	Event *EscrowParticipantCancelled // Event containing the contract specifics and raw log

	contract *bind.BoundContract // Generic contract to use for unpacking event data
	event    string              // Event name to use for unpacking event data

	logs chan types.Log        // Log channel receiving the found contract events
	sub  ethereum.Subscription // Subscription for errors, completion and termination
	done bool                  // Whether the subscription completed delivering logs
	fail error                 // Occurred error to stop iteration
}

// Next advances the iterator to the subsequent event, returning whether there
// are any more events found. In case of a retrieval or parsing error, false is
// returned and Error() can be queried for the exact failure.
func (it *EscrowParticipantCancelledIterator) Next() bool {
	// If the iterator failed, stop iterating
	if it.fail != nil {
		return false
	}
	// If the iterator completed, deliver directly whatever's available
	if it.done {
		select {
		case log := <-it.logs:
			it.Event = new(EscrowParticipantCancelled)
			if err := it.contract.UnpackLog(it.Event, it.event, log); err != nil {
				it.fail = err
				return false
			}
			it.Event.Raw = log
			return true

		default:
			return false
		}
	}
	// Iterator still in progress, wait for either a data or an error event
	select {
	case log := <-it.logs:
		it.Event = new(EscrowParticipantCancelled)
		if err := it.contract.UnpackLog(it.Event, it.event, log); err != nil {
			it.fail = err
			return false
		}
		it.Event.Raw = log
		return true

	case err := <-it.sub.Err():
		it.done = true
		it.fail = err
		return it.Next()
	}
}

// Error returns any retrieval or parsing error occurred during filtering.
func (it *EscrowParticipantCancelledIterator) Error() error {
	return it.fail
}

// Close terminates the iteration process, releasing any pending underlying
// resources.
func (it *EscrowParticipantCancelledIterator) Close() error {
	it.sub.Unsubscribe()
	return nil
}

// EscrowParticipantCancelled represents a ParticipantCancelled event raised by the Escrow contract.
type EscrowParticipantCancelled struct {
	Seq         *big.Int
	Participant common.Address
	Raw         types.Log // Blockchain specific contextual infos
}

// FilterParticipantCancelled is a free log retrieval operation binding the contract event 0x91802c983b6a8e1758dd1f19589461b58e3a7c6209dbb1b45c4f382f052a1d52.
//
// Solidity: event ParticipantCancelled(seq uint256, participant address)
func (_Escrow *EscrowFilterer) FilterParticipantCancelled(opts *bind.FilterOpts) (*EscrowParticipantCancelledIterator, error) {

	logs, sub, err := _Escrow.contract.FilterLogs(opts, "ParticipantCancelled")
	if err != nil {
		return nil, err
	}
	return &EscrowParticipantCancelledIterator{contract: _Escrow.contract, event: "ParticipantCancelled", logs: logs, sub: sub}, nil
}

// WatchParticipantCancelled is a free log subscription operation binding the contract event 0x91802c983b6a8e1758dd1f19589461b58e3a7c6209dbb1b45c4f382f052a1d52.
//
// Solidity: event ParticipantCancelled(seq uint256, participant address)
func (_Escrow *EscrowFilterer) WatchParticipantCancelled(opts *bind.WatchOpts, sink chan<- *EscrowParticipantCancelled) (event.Subscription, error) {

	logs, sub, err := _Escrow.contract.WatchLogs(opts, "ParticipantCancelled")
	if err != nil {
		return nil, err
	}
	return event.NewSubscription(func(quit <-chan struct{}) error {
		defer sub.Unsubscribe()
		for {
			select {
			case log := <-logs:
				// New log arrived, parse the event and forward to the user
				event := new(EscrowParticipantCancelled)
				if err := _Escrow.contract.UnpackLog(event, "ParticipantCancelled", log); err != nil {
					return err
				}
				event.Raw = log

				select {
				case sink <- event:
				case err := <-sub.Err():
					return err
				case <-quit:
					return nil
				}
			case err := <-sub.Err():
				return err
			case <-quit:
				return nil
			}
		}
	}), nil
}

// EscrowParticipantPaidIterator is returned from FilterParticipantPaid and is used to iterate over the raw logs and unpacked data for ParticipantPaid events raised by the Escrow contract.
type EscrowParticipantPaidIterator struct {
	Event *EscrowParticipantPaid // Event containing the contract specifics and raw log

	contract *bind.BoundContract // Generic contract to use for unpacking event data
	event    string              // Event name to use for unpacking event data

	logs chan types.Log        // Log channel receiving the found contract events
	sub  ethereum.Subscription // Subscription for errors, completion and termination
	done bool                  // Whether the subscription completed delivering logs
	fail error                 // Occurred error to stop iteration
}

// Next advances the iterator to the subsequent event, returning whether there
// are any more events found. In case of a retrieval or parsing error, false is
// returned and Error() can be queried for the exact failure.
func (it *EscrowParticipantPaidIterator) Next() bool {
	// If the iterator failed, stop iterating
	if it.fail != nil {
		return false
	}
	// If the iterator completed, deliver directly whatever's available
	if it.done {
		select {
		case log := <-it.logs:
			it.Event = new(EscrowParticipantPaid)
			if err := it.contract.UnpackLog(it.Event, it.event, log); err != nil {
				it.fail = err
				return false
			}
			it.Event.Raw = log
			return true

		default:
			return false
		}
	}
	// Iterator still in progress, wait for either a data or an error event
	select {
	case log := <-it.logs:
		it.Event = new(EscrowParticipantPaid)
		if err := it.contract.UnpackLog(it.Event, it.event, log); err != nil {
			it.fail = err
			return false
		}
		it.Event.Raw = log
		return true

	case err := <-it.sub.Err():
		it.done = true
		it.fail = err
		return it.Next()
	}
}

// Error returns any retrieval or parsing error occurred during filtering.
func (it *EscrowParticipantPaidIterator) Error() error {
	return it.fail
}

// Close terminates the iteration process, releasing any pending underlying
// resources.
func (it *EscrowParticipantPaidIterator) Close() error {
	it.sub.Unsubscribe()
	return nil
}

// EscrowParticipantPaid represents a ParticipantPaid event raised by the Escrow contract.
type EscrowParticipantPaid struct {
	Seq         *big.Int
	Participant common.Address
	Raw         types.Log // Blockchain specific contextual infos
}

// FilterParticipantPaid is a free log retrieval operation binding the contract event 0xc7d538e920165422bcfcc5aded5d1866e4ddca05191db88a2cf99b0377dcecee.
//
// Solidity: event ParticipantPaid(seq uint256, participant address)
func (_Escrow *EscrowFilterer) FilterParticipantPaid(opts *bind.FilterOpts) (*EscrowParticipantPaidIterator, error) {

	logs, sub, err := _Escrow.contract.FilterLogs(opts, "ParticipantPaid")
	if err != nil {
		return nil, err
	}
	return &EscrowParticipantPaidIterator{contract: _Escrow.contract, event: "ParticipantPaid", logs: logs, sub: sub}, nil
}

// WatchParticipantPaid is a free log subscription operation binding the contract event 0xc7d538e920165422bcfcc5aded5d1866e4ddca05191db88a2cf99b0377dcecee.
//
// Solidity: event ParticipantPaid(seq uint256, participant address)
func (_Escrow *EscrowFilterer) WatchParticipantPaid(opts *bind.WatchOpts, sink chan<- *EscrowParticipantPaid) (event.Subscription, error) {

	logs, sub, err := _Escrow.contract.WatchLogs(opts, "ParticipantPaid")
	if err != nil {
		return nil, err
	}
	return event.NewSubscription(func(quit <-chan struct{}) error {
		defer sub.Unsubscribe()
		for {
			select {
			case log := <-logs:
				// New log arrived, parse the event and forward to the user
				event := new(EscrowParticipantPaid)
				if err := _Escrow.contract.UnpackLog(event, "ParticipantPaid", log); err != nil {
					return err
				}
				event.Raw = log

				select {
				case sink <- event:
				case err := <-sub.Err():
					return err
				case <-quit:
					return nil
				}
			case err := <-sub.Err():
				return err
			case <-quit:
				return nil
			}
		}
	}), nil
}

// EscrowParticipantReleasedIterator is returned from FilterParticipantReleased and is used to iterate over the raw logs and unpacked data for ParticipantReleased events raised by the Escrow contract.
type EscrowParticipantReleasedIterator struct {
	Event *EscrowParticipantReleased // Event containing the contract specifics and raw log

	contract *bind.BoundContract // Generic contract to use for unpacking event data
	event    string              // Event name to use for unpacking event data

	logs chan types.Log        // Log channel receiving the found contract events
	sub  ethereum.Subscription // Subscription for errors, completion and termination
	done bool                  // Whether the subscription completed delivering logs
	fail error                 // Occurred error to stop iteration
}

// Next advances the iterator to the subsequent event, returning whether there
// are any more events found. In case of a retrieval or parsing error, false is
// returned and Error() can be queried for the exact failure.
func (it *EscrowParticipantReleasedIterator) Next() bool {
	// If the iterator failed, stop iterating
	if it.fail != nil {
		return false
	}
	// If the iterator completed, deliver directly whatever's available
	if it.done {
		select {
		case log := <-it.logs:
			it.Event = new(EscrowParticipantReleased)
			if err := it.contract.UnpackLog(it.Event, it.event, log); err != nil {
				it.fail = err
				return false
			}
			it.Event.Raw = log
			return true

		default:
			return false
		}
	}
	// Iterator still in progress, wait for either a data or an error event
	select {
	case log := <-it.logs:
		it.Event = new(EscrowParticipantReleased)
		if err := it.contract.UnpackLog(it.Event, it.event, log); err != nil {
			it.fail = err
			return false
		}
		it.Event.Raw = log
		return true

	case err := <-it.sub.Err():
		it.done = true
		it.fail = err
		return it.Next()
	}
}

// Error returns any retrieval or parsing error occurred during filtering.
func (it *EscrowParticipantReleasedIterator) Error() error {
	return it.fail
}

// Close terminates the iteration process, releasing any pending underlying
// resources.
func (it *EscrowParticipantReleasedIterator) Close() error {
	it.sub.Unsubscribe()
	return nil
}

// EscrowParticipantReleased represents a ParticipantReleased event raised by the Escrow contract.
type EscrowParticipantReleased struct {
	Seq         *big.Int
	Participant common.Address
	Raw         types.Log // Blockchain specific contextual infos
}

// FilterParticipantReleased is a free log retrieval operation binding the contract event 0x6f29a1913da999f186768f08d8f2ed16f36182d1e2073a748f825b38a21f7990.
//
// Solidity: event ParticipantReleased(seq uint256, participant address)
func (_Escrow *EscrowFilterer) FilterParticipantReleased(opts *bind.FilterOpts) (*EscrowParticipantReleasedIterator, error) {

	logs, sub, err := _Escrow.contract.FilterLogs(opts, "ParticipantReleased")
	if err != nil {
		return nil, err
	}
	return &EscrowParticipantReleasedIterator{contract: _Escrow.contract, event: "ParticipantReleased", logs: logs, sub: sub}, nil
}

// WatchParticipantReleased is a free log subscription operation binding the contract event 0x6f29a1913da999f186768f08d8f2ed16f36182d1e2073a748f825b38a21f7990.
//
// Solidity: event ParticipantReleased(seq uint256, participant address)
func (_Escrow *EscrowFilterer) WatchParticipantReleased(opts *bind.WatchOpts, sink chan<- *EscrowParticipantReleased) (event.Subscription, error) {

	logs, sub, err := _Escrow.contract.WatchLogs(opts, "ParticipantReleased")
	if err != nil {
		return nil, err
	}
	return event.NewSubscription(func(quit <-chan struct{}) error {
		defer sub.Unsubscribe()
		for {
			select {
			case log := <-logs:
				// New log arrived, parse the event and forward to the user
				event := new(EscrowParticipantReleased)
				if err := _Escrow.contract.UnpackLog(event, "ParticipantReleased", log); err != nil {
					return err
				}
				event.Raw = log

				select {
				case sink <- event:
				case err := <-sub.Err():
					return err
				case <-quit:
					return nil
				}
			case err := <-sub.Err():
				return err
			case <-quit:
				return nil
			}
		}
	}), nil
}

