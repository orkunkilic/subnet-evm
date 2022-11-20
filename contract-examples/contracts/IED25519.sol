// (c) 2022-2023, Ava Labs, Inc. All rights reserved.
// See the file LICENSE for licensing terms.

// SPDX-License-Identifier: MIT

pragma solidity >=0.8.0;

interface IED25519 {
  function verify(bytes32 signature, bytes32 message, bytes32 publicKey) external returns (bool);
}