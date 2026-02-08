import { useEffect, useState, useCallback } from 'react';
import useSWR, { mutate } from 'swr';
import { useTranslation } from 'react-i18next';
import {
  Flex,
  Text,
  Heading,
  Box,
  Grid,
  Container,
  Theme,
} from '@radix-ui/themes';
import { AlertCircle } from 'lucide-react';

// Types
import type {
  AuthStatus,
  Configuration,
  Instance,
  InstanceType,
} from './types';

// Components
import { LoadingScreen } from './components/common/LoadingScreen';
import { ErrorScreen } from './components/common/ErrorScreen';
import { LoginView } from './components/auth/LoginView';
import { Header } from './components/layout/Header';
import { InstanceCard } from './components/dashboard/InstanceCard';
import { CreateInstanceCard } from './components/dashboard/CreateInstanceCard';
import { CreateInstanceView } from './components/dashboard/CreateInstanceView';

const fetcher = (url: string) =>
  fetch(url).then(async (res) => {
    if (res.status === 401) return null;
    if (!res.ok) throw new Error('Failed to fetch');
    return res.json();
  });

function App() {
  const { t } = useTranslation();
  const [authError, setAuthError] = useState<string>('');
  const [isCreating, setIsCreating] = useState(false);
  const [view, setView] = useState<'dashboard' | 'create'>('dashboard');

  // Configuration
  const { data: config } = useSWR<Configuration>(
    '/_hakoniwa/api/configuration',
    fetcher,
    {
      onSuccess: (data) => {
        if (data?.title) {
          document.title = data.title;
        }
      },
    }
  );

  // Auth
  const {
    data: authData,
    error: swrAuthError,
    isLoading: isAuthLoading,
  } = useSWR<AuthStatus | null>('/_hakoniwa/api/auth/me', fetcher, {
    onError: (err) => {
      console.error(err);
      setAuthError(t('error.connection_failed'));
    },
  });

  // Instances
  const {
    data: instances,
    error: instancesError,
    isLoading: isInstancesLoading,
  } = useSWR<Instance[]>(
    authData && view === 'dashboard' ? '/_hakoniwa/api/instances' : null,
    fetcher,
    {
      refreshInterval: 3000,
    }
  );

  // Instance Types
  const { data: instanceTypes } = useSWR<InstanceType[]>(
    authData ? '/_hakoniwa/api/instance-types' : null,
    fetcher
  );

  // Create Instance Action
  const createInstance = useCallback(
    async (typeId: string, persistent: boolean = false) => {
      setIsCreating(true);
      setAuthError('');
      try {
        const res = await fetch('/_hakoniwa/api/instances', {
          method: 'POST',
          headers: {
            'Content-Type': 'application/json',
          },
          body: JSON.stringify({ type: typeId, persistent }),
        });
        if (res.status === 503) {
          throw new Error(t('error.max_instances'));
        }
        if (!res.ok) throw new Error('Failed to create instance');
        await mutate('/_hakoniwa/api/instances');
        setView('dashboard');
      } catch (err: unknown) {
        console.error(err);
        const message =
          err instanceof Error ? err.message : t('error.generic_desc');
        setAuthError(message);
      } finally {
        setIsCreating(false);
      }
    },
    [t]
  );

  // Delete Instance Action
  const deleteInstance = useCallback(async (id: string) => {
    try {
      const res = await fetch(`/_hakoniwa/api/instances/${id}`, {
        method: 'DELETE',
      });
      if (!res.ok) throw new Error('Failed to delete instance');
      await mutate('/_hakoniwa/api/instances');
    } catch (err: unknown) {
      console.error(err);
      // Optionally show error
    }
  }, []);

  // Login Anonymous Action
  const loginAnonymous = useCallback(async () => {
    setAuthError('');
    try {
      const res = await fetch('/_hakoniwa/api/auth/anonymous', {
        method: 'POST',
      });
      if (!res.ok) throw new Error('Login failed');
      // After login, force revalidate /auth/me
      await mutate('/_hakoniwa/api/auth/me');
    } catch (err) {
      console.error(err);
      setAuthError(t('error.login_failed'));
    }
  }, [t]);

  // Logout Action
  const logout = useCallback(async () => {
    try {
      await fetch('/_hakoniwa/api/auth/logout', {
        method: 'POST',
      });
      // Clear client-side cookies just in case, though backend should handle it
      document.cookie =
        'hakoniwa_session=; path=/; expires=Thu, 01 Jan 1970 00:00:00 GMT';
      document.cookie =
        'hakoniwa_instance_id=; path=/; expires=Thu, 01 Jan 1970 00:00:00 GMT';
      window.location.reload();
    } catch (err) {
      console.error('Logout failed', err);
      window.location.reload();
    }
  }, []);

  // Auto Login Check
  useEffect(() => {
    if (
      isAuthLoading ||
      !config ||
      !config.auth_auto_login ||
      authError ||
      authData
    )
      return;

    // Don't skip if we are already in an error state from URL
    if (new URLSearchParams(window.location.search).get('error')) return;

    const methods = config.auth_methods;
    if (methods.length === 1) {
      const method = methods[0];
      if (method === 'oidc') {
        window.location.href = '/_hakoniwa/api/auth/oidc/authorize';
      } else if (method === 'anonymous') {
        loginAnonymous();
      }
    }
  }, [isAuthLoading, config, authError, authData, loginAnonymous]);

  // Check for errors in URL (redirected from backend)
  useEffect(() => {
    const params = new URLSearchParams(window.location.search);
    const error = params.get('error');
    if (error) {
      console.error('Auth error:', error);
      setAuthError(t('error.login_failed'));
      // Clean URL
      window.history.replaceState({}, document.title, window.location.pathname);
    }
  }, [t]);

  const isLoading =
    isAuthLoading || (authData && isInstancesLoading && view === 'dashboard');

  // 1. Loading State
  if (isLoading && !authError) {
    return <LoadingScreen title={config?.title} />;
  }

  // 2. Authenticated View
  if (authData) {
    return (
      <Theme accentColor="indigo" grayColor="slate" radius="large">
        <Box
          style={{
            minHeight: '100vh',
            backgroundColor: 'var(--gray-2)',
            backgroundImage:
              'radial-gradient(circle at 50% 0, var(--indigo-3), transparent 50%)',
          }}>
          <Header config={config} user={authData.user} onLogout={logout} />

          {/* Main Content */}
          <Container size="4" p="4" pt="8">
            {instancesError && (
              <Box
                mb="4"
                p="3"
                style={{
                  background: 'var(--red-3)',
                  borderRadius: 'var(--radius-2)',
                }}>
                <Flex gap="2" align="center">
                  <AlertCircle size={16} color="var(--red-11)" />
                  <Text color="red" size="2" weight="medium">
                    {t('error.generic_desc')}
                  </Text>
                </Flex>
              </Box>
            )}
            {view === 'dashboard' ? (
              <Flex direction="column" gap="6">
                <Flex justify="between" align="end">
                  <Box>
                    <Heading
                      size="8"
                      mb="2"
                      style={{ fontWeight: 900, letterSpacing: '-0.03em' }}>
                      {t('dashboard.title')}
                    </Heading>
                    <Text size="3" color="gray">
                      {t('dashboard.subtitle')}
                    </Text>
                  </Box>
                </Flex>

                <Grid
                  columns={{ initial: '1', sm: '2', lg: '3', xl: '4' }}
                  gap="6">
                  {/* Instance Cards */}
                  {instances?.map((inst, index) => {
                    const typeInfo = instanceTypes?.find(
                      (t) => t.id === inst.type
                    );
                    return (
                      <InstanceCard
                        key={inst.id}
                        instance={inst}
                        typeInfo={typeInfo}
                        index={index}
                        onDelete={deleteInstance}
                        onOpen={(id) => {
                          document.cookie = `hakoniwa_instance_id=${id}; path=/`;
                          window.location.href = '/';
                        }}
                      />
                    );
                  })}

                  {/* Create New Card Button */}
                  <CreateInstanceCard
                    delayIndex={instances?.length || 0}
                    onClick={() => setView('create')}
                  />
                </Grid>
              </Flex>
            ) : (
              /* Create Workspace Screen */
              <CreateInstanceView
                user={authData.user}
                instanceTypes={instanceTypes}
                enablePersistence={config?.enable_persistence ?? false}
                isCreating={isCreating}
                error={authError}
                onBack={() => setView('dashboard')}
                onCreate={createInstance}
              />
            )}
          </Container>
        </Box>
      </Theme>
    );
  }

  if (authError || swrAuthError) {
    return (
      <ErrorScreen error={authError} onRetry={() => window.location.reload()} />
    );
  }

  // 3. Login View (Unauthenticated)
  return <LoginView config={config} onLoginAnonymous={loginAnonymous} />;
}

export default App;
