import { Flex, Card, Box, Heading, Text, Button, Link as RadixLink } from '@radix-ui/themes';
import { ArrowRight } from 'lucide-react';
import { useTranslation, Trans } from 'react-i18next';
import type { Configuration } from '../../types';

interface LoginViewProps {
  config?: Configuration;
  onLoginAnonymous: () => void;
}

export function LoginView({ config, onLoginAnonymous }: LoginViewProps) {
  const { t } = useTranslation();

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
            {config?.auth_methods.includes('oidc') && (
              <Button
                size="3"
                onClick={() => {
                  window.location.href = '/_hakoniwa/api/auth/oidc/authorize';
                }}
                style={{ height: '48px', fontSize: '16px', cursor: 'pointer' }}
                className="login-button">
                {t('login.oidc_button', {
                  name: config?.oidc_name || 'OpenID Connect',
                })}
                <ArrowRight className="login-button-arrow" />
              </Button>
            )}

            {config?.auth_methods.includes('oidc') &&
              config?.auth_methods.includes('anonymous') && (
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
              )}

            {config?.auth_methods.includes('anonymous') && (
              <Button
                size="3"
                onClick={onLoginAnonymous}
                style={{ height: '48px', fontSize: '16px', cursor: 'pointer' }}
                className="login-button">
                {t('login.anonymous_button')}
                <ArrowRight className="login-button-arrow" />
              </Button>
            )}
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
