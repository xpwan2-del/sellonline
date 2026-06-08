const currentBranch = process.env.VERCEL_GIT_COMMIT_REF || ''
const deployBranch = process.env.DEPLOY_BRANCH || ''

if (!deployBranch) {
  console.log('DEPLOY_BRANCH is not set; skipping deployment to avoid cross-environment updates.')
  process.exit(0)
}

if (currentBranch !== deployBranch) {
  console.log(`Skipping deployment for branch "${currentBranch}". Expected "${deployBranch}".`)
  process.exit(0)
}

console.log(`Deploying branch "${currentBranch}" for DEPLOY_BRANCH="${deployBranch}".`)
process.exit(1)
