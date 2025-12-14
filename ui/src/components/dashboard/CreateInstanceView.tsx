import { useState } from 'react';
import { Flex, Box, Heading, Text, Grid, Card, AspectRatio, Button, Checkbox, Tooltip } from '@radix-ui/themes';
import { ArrowLeft, AlertCircle, Terminal, CheckCircle2, HardDrive } from 'lucide-react';
import { useTranslation } from 'react-i18next';
import type { InstanceType, User } from '../../types';

interface CreateInstanceViewProps {
  user?: User;
  instanceTypes?: InstanceType[];
  enablePersistence: boolean;
  isCreating: boolean;
  error?: string;
  onBack: () => void;
  onCreate: (typeId: string, persistent: boolean) => void;
}

export function CreateInstanceView({ user, instanceTypes, enablePersistence, isCreating, error, onBack, onCreate }: CreateInstanceViewProps) {
  const { t } = useTranslation();
  const [selectedTypeId, setSelectedTypeId] = useState<string | null>(null);
  const [persistent, setPersistent] = useState(false);

  const selectedType = instanceTypes?.find(t => t.id === selectedTypeId);
  const canUsePersistent = enablePersistence && user?.type === 'openid_connect' && (selectedType?.persistable ?? false);

  const getPersistentDisabledReason = () => {
    if (!enablePersistence) return t('workspace.create.persistent_disabled_global');
    if (user?.type !== 'openid_connect') return t('workspace.create.persistent_disabled_hint');
    if (selectedType && !selectedType.persistable) return t('workspace.create.persistent_disabled_type');
    return t('workspace.create.persistent_disabled_hint');
  };

  return (
    <Flex direction="column" gap="6" className="anim-entry" style={{ height: '100%' }}>
      {/* Header Section */}
      <Flex direction="column" gap="4">
        <Button variant="ghost" color="gray" onClick={onBack} style={{ alignSelf: 'start', padding: 0, height: 'auto', gap: '8px' }}>
          <ArrowLeft size={20} />
          {t('workspace.action.back_to_list')}
        </Button>
        <Box>
          <Heading size="8" weight="bold" style={{ letterSpacing: '-0.03em', marginBottom: '12px' }}>{t('workspace.create.modal_title')}</Heading>
          <Text size="3" color="gray">{t('workspace.create.modal_desc')}</Text>
        </Box>
      </Flex>

      {error && (
        <Box p="3" style={{ background: 'var(--red-3)', borderRadius: 'var(--radius-2)' }}>
          <Flex gap="2" align="center">
            <AlertCircle size={16} color="var(--red-11)" />
            <Text color="red" size="2" weight="medium">{error}</Text>
          </Flex>
        </Box>
      )}

      {/* Selection Grid */}
      <Grid columns={{ initial: '1', sm: '2', md: '3' }} gap="4">
        {instanceTypes?.map((type) => {
          const isSelected = selectedTypeId === type.id;
          return (
            <Box 
              key={type.id}
              onClick={() => {
                 setSelectedTypeId(type.id);
                 if (!type.persistable) setPersistent(false);
              }}
              style={{ 
                cursor: isCreating ? 'wait' : 'pointer',
                position: 'relative',
                transition: 'all 0.2s ease'
              }}
            >
              <Card 
                size="3"
                style={{ 
                  padding: '16px',
                  border: isSelected ? '2px solid var(--accent-9)' : '2px solid transparent',
                  backgroundColor: isSelected ? 'var(--accent-2)' : 'var(--gray-2)',
                  height: '100%',
                }}
                className={isSelected ? 'selected-card' : 'hover-card'}
              >
                <Flex direction="column" gap="3">
                  <AspectRatio ratio={16/10}>
                    <Box 
                      style={{ 
                        width: '100%', 
                        height: '100%', 
                        background: 'var(--gray-4)',
                        borderRadius: 'var(--radius-3)',
                        overflow: 'hidden',
                        display: 'flex',
                        alignItems: 'center',
                        justifyContent: 'center',
                        position: 'relative'
                      }}
                    >
                      {type.logo_url ? (
                        <img 
                          src={type.logo_url} 
                          alt={type.name} 
                          style={{ width: '100%', height: '100%', objectFit: 'cover' }} 
                        />
                      ) : (
                        <Terminal size={48} color="var(--gray-8)" />
                      )}
                      
                      {isSelected && (
                        <Box 
                          style={{
                            position: 'absolute',
                            top: 8,
                            right: 8,
                            background: 'var(--accent-9)',
                            borderRadius: '50%',
                            padding: 4,
                            color: 'white',
                            boxShadow: '0 2px 4px rgba(0,0,0,0.1)',
                            lineHeight: 0,
                          }}
                        >
                          <CheckCircle2 size={16} />
                        </Box>
                      )}
                    </Box>
                  </AspectRatio>
                  
                  <Box>
                    <Heading size="4" mb="1" weight="medium">{type.name}</Heading>
                    {type.description && (
                      <Text size="2" color="gray" style={{ lineHeight: '1.5' }}>
                        {type.description}
                      </Text>
                    )}
                  </Box>
                </Flex>
              </Card>
            </Box>
          );
        })}
      </Grid>

      {/* Footer Action */}
      <Box 
        mt="auto" 
        pt="4" 
        style={{ 
          borderTop: '1px solid var(--gray-4)',
          display: 'flex',
          justifyContent: 'flex-end',
          alignItems: 'center',
          gap: '16px'
        }}
      >
        {user && (
          <Flex align="center" gap="2">
            {canUsePersistent ? (
               <Box style={{ opacity: 1 }}>
                 <Flex asChild align="center" gap="2">
                    <label style={{ cursor: 'pointer', display: 'flex', alignItems: 'center', gap: '8px' }}>
                        <Checkbox 
                          checked={persistent} 
                          onCheckedChange={(c) => setPersistent(!!c)} 
                        />
                        <Flex gap="1" align="center">
                          <HardDrive size={16} />
                          <Text size="2">{t('workspace.create.enable_persistent')}</Text>
                        </Flex>
                    </label>
                 </Flex>
               </Box>
            ) : (
              <Tooltip content={getPersistentDisabledReason()}>
                 <Box style={{ opacity: 0.5 }}>
                   <Flex asChild align="center" gap="2">
                      <label style={{ cursor: 'not-allowed', display: 'flex', alignItems: 'center', gap: '8px' }}>
                          <Checkbox 
                            checked={persistent} 
                            disabled={true} 
                          />
                          <Flex gap="1" align="center">
                            <HardDrive size={16} />
                            <Text size="2">{t('workspace.create.enable_persistent')}</Text>
                          </Flex>
                      </label>
                   </Flex>
                 </Box>
              </Tooltip>
            )}
          </Flex>
        )}

        <Button 
          size="4" 
          disabled={!selectedTypeId || isCreating} 
          onClick={() => selectedTypeId && onCreate(selectedTypeId, persistent)}
          style={{ 
            width: '100%', 
            maxWidth: '240px', 
            fontWeight: 'bold' 
          }}
        >
          {isCreating ? t('workspace.status.creating') : t('workspace.create.submit')}
        </Button>
      </Box>
    </Flex>
  );
}
