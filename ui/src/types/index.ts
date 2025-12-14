export type InstanceStatus = 'pending' | 'running' | 'terminating';

export interface Instance {
  id: string;
  name: string;
  type: string;
  status: InstanceStatus;
  pod_ip?: string;
}

export interface InstanceType {
  id: string;
  name: string;
  description?: string;
  logo_url?: string;
  persistable?: boolean;
}

export interface User {
  id: string;
  type: 'openid_connect' | 'anonymous';
}

export interface AuthStatus {
  user: User;
}

export interface Configuration {
  title: string;
  message: string;
  logo_url: string;
  terms_of_service_url?: string;
  privacy_policy_url?: string;
  auth_methods: string[];
  oidc_name: string;
  auth_auto_login: boolean;
  enable_persistence: boolean;
}

export interface CreateInstancePayload {
  type: string;
  persistent?: boolean;
}
