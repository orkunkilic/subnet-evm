// (c) 2022-2023, Ava Labs, Inc. All rights reserved.
// See the file LICENSE for licensing terms.

// SPDX-License-Identifier: MIT

pragma solidity >=0.8.0;

interface IContractCallMainnet {
    function call(address _contract, bytes calldata _data) external returns (bytes memory);
} 