# Testing Strategy

## Test-Case Identification

During backlog refinement, the acceptance criteria for new features are identified, which later become test cases for functional tests. 

## Functional Tests

 - Unit tests, [executed on every pull request](../../.github/workflows/pull.yaml), test the individual units of application logic in isolation, focusing on the implementation correctness. 
 - Integration tests, executed with every [push to main branch]((../../.github/workflows/push.yaml)) and [release](../../.github/workflows/release.yaml) test the main usage scenario of routing image pull requests via node port and reverse proxy workload in a cluster and verifies the functional correctness of the component. 
 - Manual E2E integration test with SAP BTP Cloud Connector is executed against the released version before enabling it for customers. For more details, see [Manual Test for Registry Proxy](manual-testing.md). This test evaluates the component on the DEV landscape using the connectivity-proxy module and a configured Cloud Connector on the SAP BTP subaccount. It also involves testing on an SKR cluster, which is a Kyma runtime cluster managed by SAP.

## Non-Functional Tests

 - Security tests provide automated scanning by Mend, BDBA, and Checkmarx 

## Source Code Quality

Source code quality prerequisites are checked automatically on every [pull request](../../.github/workflows/lint.yaml)
This workflow performs static code checks, including formatting, linting, and complexity analysis, using the official [golangci-lint GH action](https://github.com/golangci/golangci-lint-action).
