@Library(['irpa-lib']) _

pipeline {
    agent {
        kubernetes {
            defaultContainer 'go-tools'
            yaml k8sConfigTemplate(['go-tools': 'v3'])
            // yaml kubeConfig
        }
    }

    environment {
        TF_ACC=1
        CF_SKIP_SSL_VALIDATION="true"
        
        TEST_ORG_NAME="IPA-CloudOps_ipa-cloudops--infra"
        TEST_SPACE_NAME="main"
        CF_CA_CERT=""
        CF_API_URL="https://api.cf.eu12.hana.ondemand.com"
        TEST_SERVICE_1="destination"
        TEST_SERVICE_2="destination"
        TEST_SERVICE_PLAN="lite"
        TEST_APP_DOMAIN="cfapps.eu12.hana.ondemand.com"
        
        logfile="acc_test.log"
    }
    stages {
        stage('Run tests') {
            steps {
                withCredentials([
                    usernamePassword(credentialsId: 'btp_tech_user', usernameVariable: 'CF_USER', passwordVariable: 'CF_PASSWORD')
                ]) {
                    sh'''#!/bin/bash
                        go version
                        export TF_LOG=info
                        go test -timeout 2400s -run '^(TestAccResAppVersions_app1|TestAccDefaultValues_app1|TestAccDefaultValuesRolling_app1|TestAccResApp_app1|TestAccResApp_Routes_updateToAndmore|TestAccResApp_dockerApp|TestAccResApp_dockerAppInvocationTimeout|TestAccResApp_app_bluegreen)$' github.com/terraform-providers/terraform-provider-cloudfoundry/cloudfoundry > ${logfile}
                        cat ${logfile}
                    '''
                
                    archiveArtifacts artifacts: "${logfile}", fingerprint: true
                }
            }
        }
    }
}
