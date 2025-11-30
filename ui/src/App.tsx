import { useEffect, useState } from 'react';
import useSWR, { mutate } from 'swr';
import { useTranslation, Trans } from 'react-i18next';
import { Loader2, AlertCircle, ArrowRight } from 'lucide-react';
import {
  Button,
  Card,
  Flex,
  Text,
  Heading,
  Box,
  Link as RadixLink,
} from '@radix-ui/themes';

type InstanceStatus = 'pending' | 'running' | 'terminating';

interface User {
  id: string;
  type: 'openid_connect' | 'anonymous';
}

interface AuthStatus {
  user: User;
  instance?: {
    status: InstanceStatus;
    pod_ip?: string;
  };
}

interface Configuration {
  title: string;
  message: string;
  logo_url: string;
  terms_of_service_url?: string;
  privacy_policy_url?: string;
}

const fetcher = (url: string) =>
  fetch(url).then(async (res) => {
    if (res.status === 401) return null;
    if (!res.ok) throw new Error('Failed to fetch');
    return res.json();
  });

function App() {
  const { t } = useTranslation();
  const [shouldPoll, setShouldPoll] = useState(false);
  const [authError, setAuthError] = useState<string>('');

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
    isLoading,
  } = useSWR<AuthStatus | null>('/_hakoniwa/api/auth/me', fetcher, {
    refreshInterval: shouldPoll ? 3000 : 0,
    onError: (err) => {
      console.error(err);
      setAuthError(t('error.connection_failed'));
    },
  });

  const instanceStatus = authData?.instance?.status;

  useEffect(() => {
    if (swrAuthError) return;

    if (instanceStatus === 'running') {
      window.location.reload();
      return;
    }

    if (instanceStatus === 'pending' || instanceStatus === 'terminating') {
      setShouldPoll(true);
    } else {
      setShouldPoll(false);
    }
  }, [instanceStatus, swrAuthError]);

  const loginAnonymous = async () => {
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
  };

  // Determine UI State
  if (isLoading && !authData && !authError) {
    return (
      <Flex align="center" justify="center" style={{ minHeight: '100vh' }}>
        <Flex direction="column" align="center" gap="4">
          <Loader2
            className="animate-spin"
            size={40}
            style={{ color: 'var(--gray-10)' }}
          />
          <Text color="gray">
            {t('status.connecting', { title: config?.title || 'Hakoniwa' })}
          </Text>
        </Flex>
      </Flex>
    );
  }

  if (
    shouldPoll ||
    (authData?.instance && authData.instance.status !== 'running')
  ) {
    return (
      <Flex
        align="center"
        justify="center"
        style={{ minHeight: '100vh' }}
        p="4">
        <Card size="3" style={{ width: '100%', maxWidth: 400 }}>
          <Flex direction="column" gap="5" align="center" py="4">
            <Box>
              <Heading size="5" align="center">
                {t('status.preparing')}
              </Heading>
              <Text color="gray" size="2" align="center" as="p">
                {t('status.spinning_up')}
              </Text>
            </Box>

            <Loader2
              className="animate-spin"
              size={64}
              style={{ color: 'var(--accent-9)' }}
            />

            <Text size="2" color="gray">
              {t('status.label')}{' '}
              <Text weight="bold" highContrast>
                {instanceStatus || t('status.unknown')}
              </Text>
            </Text>
          </Flex>
        </Card>
      </Flex>
    );
  }

  if (authError || swrAuthError) {
    return (
      <Flex
        align="center"
        justify="center"
        style={{ minHeight: '100vh' }}
        p="4">
        <Card size="3" style={{ width: '100%', maxWidth: 400 }}>
          <Flex gap="4" direction="column">
            <Flex gap="2" align="center">
              <AlertCircle color="var(--red-9)" />
              <Heading size="4" color="red">
                {t('error.title')}
              </Heading>
            </Flex>
            <Text size="2" color="gray">
              {t('error.generic_desc')}
            </Text>

            <Box
              p="3"
              style={{
                background: 'var(--gray-3)',
                borderRadius: 'var(--radius-2)',
              }}>
              <Text size="2" color="red">
                {authError || t('error.connection_failed')}
              </Text>
            </Box>

            <Button
              color="red"
              variant="soft"
              onClick={() => window.location.reload()}
              style={{ width: '100%' }}>
              {t('action.retry')}
            </Button>
          </Flex>
        </Card>
      </Flex>
    );
  }

  // Idle state (Not logged in)
  return (
    <Flex
      align="center"
      justify="center"
      style={{
        minHeight: '100vh',
        background:
          'linear-gradient(to bottom right, var(--blue-9), var(--purple-9))',
      }}
      p="4">
      <Card
        size="4"
        style={{
          width: '100%',
          maxWidth: 480,
          boxShadow:
            '0 10px 25px -5px rgba(0, 0, 0, 0.1), 0 8px 10px -6px rgba(0, 0, 0, 0.1)',
        }}>
        <Flex direction="column" gap="6" py="2">
          <Flex direction="column" align="center" gap="4">
            <Box
              style={{
                width: '120px',
                height: '120px',
                display: 'flex',
                alignItems: 'center',
                justifyContent: 'center',
              }}>
              <img
                src={config?.logo_url || '/_hakoniwa/hakoniwa_logo.webp'}
                alt="Logo"
                style={{
                  maxWidth: '100%',
                  maxHeight: '100%',
                  objectFit: 'contain',
                }}
              />
            </Box>
            <Box>
              <Heading size="8" align="center" style={{ marginBottom: '8px' }}>
                {config?.title || 'Hakoniwa'}
              </Heading>
              {config?.message && (
                <Text size="4" color="gray" align="center" as="p">
                  {config?.message}
                </Text>
              )}
            </Box>
          </Flex>

          <Flex direction="column" gap="3" mt="2">
            <Button
              size="3"
              variant="outline"
              disabled
              style={{ height: '48px', fontSize: '16px' }}>
              {t('login.oidc_button')}
            </Button>

            <Flex align="center" gap="2">
              <Box
                style={{ flex: 1, height: 1, background: 'var(--gray-5)' }}
              />
              <Text
                size="1"
                color="gray"
                style={{ textTransform: 'uppercase' }}>
                {t('login.or_continue')}
              </Text>
              <Box
                style={{ flex: 1, height: 1, background: 'var(--gray-5)' }}
              />
            </Flex>

            <Button
              size="3"
              onClick={loginAnonymous}
              style={{ height: '48px', fontSize: '16px', cursor: 'pointer' }}
              className="anonymous-button">
              {t('login.anonymous_button')}
              <ArrowRight className="anonymous-button-arrow" />
            </Button>
          </Flex>

          {(config?.terms_of_service_url || config?.privacy_policy_url) && (
            <Text size="1" color="gray" align="center" mt="2">
              {config.terms_of_service_url && config.privacy_policy_url ? (
                <Trans
                  i18nKey="legal.agreement"
                  components={[
                    <RadixLink
                      href={config.terms_of_service_url}
                      target="_blank"
                      rel="noopener noreferrer"
                    />,
                    <RadixLink
                      href={config.privacy_policy_url}
                      target="_blank"
                      rel="noopener noreferrer"
                    />,
                  ]}
                />
              ) : config.terms_of_service_url ? (
                <Trans
                  i18nKey="legal.agreement_tos_only"
                  components={[
                    <RadixLink
                      href={config.terms_of_service_url}
                      target="_blank"
                      rel="noopener noreferrer"
                    />,
                  ]}
                />
              ) : config.privacy_policy_url ? (
                <Trans
                  i18nKey="legal.agreement_privacy_only"
                  components={[
                    <RadixLink
                      href={config.privacy_policy_url}
                      target="_blank"
                      rel="noopener noreferrer"
                    />,
                  ]}
                />
              ) : null}
            </Text>
          )}
        </Flex>
      </Card>
    </Flex>
  );
}

export default App;
