export function AddSecret(...secrets) {
  if (!secrets || secrets.length < 2) {
    secrets = ['','']
  }
  global['ops-project-name'] = secrets[0]
  global['ops-secret'] = secrets[1]
}
