/* Auto-generated. DO NOT EDIT. Generator: https://golift.io/goty
 * Edit the source code and run goty again to make updates.
 */

/**
 * The day of the week.
 * @see golang: <time.Weekday>
 */
export enum Weekday {
  Sunday    = 0,
  Monday    = 1,
  Tuesday   = 2,
  Wednesday = 3,
  Thursday  = 4,
  Friday    = 5,
  Saturday  = 6,
};

/**
 * Config represents the data in our config file.
 * @see golang: <github.com/Notifiarr/notifiarr/pkg/configfile.Config>
 */
export interface NotiConfig extends NotiLogConfig, NotiApps {
  hostId: string;
  uiPassword: string;
  bindAddr: string;
  sslCertFile: string;
  sslKeyFile: string;
  upstreams?: string[];
  autoUpdate: string;
  unstableCh: boolean;
  timeout: string;
  retries: number;
  snapshot?: NotiSnapshotConfig;
  services?: NotiServicesConfig;
  service?: (null | NotiService)[];
  apt: boolean;
  watchFiles?: (null | NotiWatchFile)[];
  endpoints?: (null | NotiEndpoint)[];
  commands?: (null | NotiCommand)[];
};

/**
 * Config determines which checks to run, etc.
 * @see golang: <github.com/Notifiarr/notifiarr/pkg/snapshot.Config>
 */
export interface NotiSnapshotConfig extends NotiPlugins {
  timeout: string;
  interval: string;
  zfsPools?: string[];
  useSudo: boolean;
  monitorRaid: boolean;
  monitorDrives: boolean;
  monitorSpace: boolean;
  allDrives: boolean;
  quotas: boolean;
  ioTop: number;
  psTop: number;
  myTop: number;
  ipmi: boolean;
  ipmiSudo: boolean;
};

/**
 * Plugins is optional configuration for "plugins".
 * @see golang: <github.com/Notifiarr/notifiarr/pkg/snapshot.Plugins>
 */
export interface NotiPlugins {
  nvidia?: NotiNvidiaConfig;
  mysql?: (null | NotiMySQLConfig)[];
};

/**
 * NvidiaConfig is our input data.
 * @see golang: <github.com/Notifiarr/notifiarr/pkg/snapshot.NvidiaConfig>
 */
export interface NotiNvidiaConfig {
  smiPath: string;
  busIDs?: string[];
  disabled: boolean;
};

/**
 * MySQLConfig allows us to gather a process list for the snapshot.
 * @see golang: <github.com/Notifiarr/notifiarr/pkg/snapshot.MySQLConfig>
 */
export interface NotiMySQLConfig {
  name: string;
  host: string;
  timeout: string;
  /**
   * Only used by service checks, snapshot interval is used for mysql.
   */
  interval: string;
};

/**
 * Config for this Services plugin comes from a config file.
 * @see golang: <github.com/Notifiarr/notifiarr/pkg/services.Config>
 */
export interface NotiServicesConfig {
  interval: string;
  parallel: number;
  disabled: boolean;
  logFile: string;
};

/**
 * Service is a thing we check and report results for.
 * @see golang: <github.com/Notifiarr/notifiarr/pkg/services.Service>
 */
export interface NotiService {
  name: string;
  type: string;
  value: string;
  expect: string;
  timeout: string;
  interval: string;
  tags?: Record<string, null | any>;
};

/**
 * WatchFile is the input data needed to watch files.
 * @see golang: <github.com/Notifiarr/notifiarr/pkg/triggers/filewatch.WatchFile>
 */
export interface NotiWatchFile {
  path: string;
  regex: string;
  skip: string;
  poll: boolean;
  pipe: boolean;
  mustExist: boolean;
  logMatch: boolean;
};

/**
 * Endpoint contains the cronjob definition and url query parameters.
 * This is the input data to poll a url on a frequency.
 * @see golang: <github.com/Notifiarr/notifiarr/pkg/triggers/endpoints/epconfig.Endpoint>
 */
export interface NotiEndpoint extends NotiCronJob {
  query?: Record<string, null | string[]>;
  header?: Record<string, null | string[]>;
  template: string;
  name: string;
  url: string;
  method: string;
  body: string;
  follow: boolean;
};

/**
 * CronJob defines when a job should run.
 * When Frequency is set to:
 * 0 `DeadCron` disables the schedule.
 * 1 `Minutely` uses Seconds.
 * 2 `Hourly` uses Minutes and Seconds.
 * 3 `Daily` uses Hours, Minutes and Seconds.
 * 4 `Weekly` uses DaysOfWeek, Hours, Minutes and Seconds.
 * 5 `Monthly` uses DaysOfMonth, Hours, Minutes and Seconds.
 * @see golang: <github.com/Notifiarr/notifiarr/pkg/triggers/common/scheduler.CronJob>
 */
export interface NotiCronJob {
  /**
   * Frequency to configure the job. Pass 0 disable the cron.
   */
  frequency: number;
  /**
   * Interval for Daily, Weekly and Monthly Frequencies. 1 = every day/week/month, 2 = every other, and so on.
   */
  interval: number;
  /**
   * AtTimes is a list of 'hours, minutes, seconds' to schedule for Daily/Weekly/Monthly frequencies.
   * Also used in Minutely and Hourly schedules, a bit awkwardly.
   */
  atTimes?: (null | number[])[];
  /**
   * DaysOfWeek is a list of days to schedule. 0-6. 0 = Sunday.
   */
  daysOfWeek?: Weekday[];
  /**
   * DaysOfMonth is a list of days to schedule. 1 to 31 or -31 to -1 to count backward.
   */
  daysOfMonth?: number[];
  /**
   * Months to schedule. 1 to 12. 1 = January.
   */
  months?: number[];
};

/**
 * Command contains the input data for a defined command.
 * It also contains some saved data about the command being run.
 * @see golang: <github.com/Notifiarr/notifiarr/pkg/triggers/commands.Command>
 */
export interface NotiCommand extends NotiCmdconfigConfig {};

/**
 * @see golang: <github.com/Notifiarr/notifiarr/pkg/triggers/commands/cmdconfig.Config>
 */
export interface NotiCmdconfigConfig {
  name: string;
  hash: string;
  shell: boolean;
  log: boolean;
  notify: boolean;
  args: number;
};

/**
 * LogConfig allows sending logs to rotating files.
 * Setting an AppName will force log creation even if LogFile and HTTPLog are empty.
 * @see golang: <github.com/Notifiarr/notifiarr/pkg/logs.LogConfig>
 */
export interface NotiLogConfig {
  logFile: string;
  debugLog: string;
  httpLog: string;
  logFiles: number;
  logFileMb: number;
  fileMode: number;
  debug: boolean;
  quiet: boolean;
  noUploads: boolean;
};

/**
 * Apps is the input configuration to relay requests to Starr apps.
 * @see golang: <github.com/Notifiarr/notifiarr/pkg/apps.Apps>
 */
export interface NotiApps {
  apiKey: string;
  extraKeys?: string[];
  urlbase: string;
  maxBody: number;
  serial: boolean;
  sonarr?: (null | NotiSonarrConfig)[];
  radarr?: (null | NotiRadarrConfig)[];
  lidarr?: (null | NotiLidarrConfig)[];
  readarr?: (null | NotiReadarrConfig)[];
  prowlarr?: (null | NotiProwlarrConfig)[];
  deluge?: (null | NotiDelugeConfig)[];
  qbit?: (null | NotiQbitConfig)[];
  rtorrent?: (null | NotiRtorrentConfig)[];
  sabnzbd?: (null | NotiSabNZBConfig)[];
  nzbget?: (null | NotiNZBGetConfig)[];
  transmission?: (null | NotiXmissionConfig)[];
  tautulli?: NotiTautulliConfig;
  plex?: NotiPlexConfig;
};

/**
 * SonarrConfig represents the input data for a Sonarr server.
 * @see golang: <github.com/Notifiarr/notifiarr/pkg/apps.SonarrConfig>
 */
export interface NotiSonarrConfig extends NotiExtraConfig, NotiStarrConfig {};

/**
 * @see golang: <github.com/Notifiarr/notifiarr/pkg/apps.ExtraConfig>
 */
export interface NotiExtraConfig {
  name: string;
  timeout: string;
  interval: string;
  validSsl: boolean;
  deletes: number;
};

/**
 * Config is the data needed to poll Radarr or Sonarr or Lidarr or Readarr.
 * At a minimum, provide a URL and API Key.
 * HTTPUser and HTTPPass are used for Basic HTTP auth, if enabled (not common).
 * Username and Password are for non-API paths with native authentication enabled.
 * @see golang: <golift.io/starr.Config>
 */
export interface NotiStarrConfig {
  apiKey: string;
  url: string;
  httpPass: string;
  httpUser: string;
  username: string;
  password: string;
};

/**
 * RadarrConfig represents the input data for a Radarr server.
 * @see golang: <github.com/Notifiarr/notifiarr/pkg/apps.RadarrConfig>
 */
export interface NotiRadarrConfig extends NotiExtraConfig, NotiStarrConfig {};

/**
 * LidarrConfig represents the input data for a Lidarr server.
 * @see golang: <github.com/Notifiarr/notifiarr/pkg/apps.LidarrConfig>
 */
export interface NotiLidarrConfig extends NotiExtraConfig, NotiStarrConfig {};

/**
 * ReadarrConfig represents the input data for a Readarr server.
 * @see golang: <github.com/Notifiarr/notifiarr/pkg/apps.ReadarrConfig>
 */
export interface NotiReadarrConfig extends NotiExtraConfig, NotiStarrConfig {};

/**
 * ProwlarrConfig represents the input data for a Prowlarr server.
 * @see golang: <github.com/Notifiarr/notifiarr/pkg/apps.ProwlarrConfig>
 */
export interface NotiProwlarrConfig extends NotiExtraConfig, NotiStarrConfig {};

/**
 * @see golang: <github.com/Notifiarr/notifiarr/pkg/apps.DelugeConfig>
 */
export interface NotiDelugeConfig extends NotiExtraConfig, NotiDelugeConfig0 {};

/**
 * Config is the data needed to poll Deluge.
 * @see golang: <golift.io/deluge.Config>
 */
export interface NotiDelugeConfig0 {
  url: string;
  password: string;
  httppass: string;
  httpuser: string;
  version: string;
};

/**
 * @see golang: <github.com/Notifiarr/notifiarr/pkg/apps.QbitConfig>
 */
export interface NotiQbitConfig extends NotiExtraConfig, NotiQbitConfig0 {};

/**
 * Config is the input data needed to return a Qbit struct.
 * This is setup to allow you to easily pass this data in from a config file.
 * @see golang: <golift.io/qbit.Config>
 */
export interface NotiQbitConfig0 {
  url: string;
  user: string;
  pass: string;
  httppass: string;
  httpuser: string;
};

/**
 * @see golang: <github.com/Notifiarr/notifiarr/pkg/apps.RtorrentConfig>
 */
export interface NotiRtorrentConfig extends NotiExtraConfig {
  url: string;
  user: string;
  pass: string;
};

/**
 * @see golang: <github.com/Notifiarr/notifiarr/pkg/apps.SabNZBConfig>
 */
export interface NotiSabNZBConfig extends NotiExtraConfig, NotiSabnzbdConfig {};

/**
 * @see golang: <github.com/Notifiarr/notifiarr/pkg/apps/apppkg/sabnzbd.Config>
 */
export interface NotiSabnzbdConfig {
  url: string;
  apiKey: string;
};

/**
 * @see golang: <github.com/Notifiarr/notifiarr/pkg/apps.NZBGetConfig>
 */
export interface NotiNZBGetConfig extends NotiExtraConfig, NotiNzbgetConfig {};

/**
 * Config is the input data needed to return a NZBGet struct.
 * This is setup to allow you to easily pass this data in from a config file.
 * @see golang: <golift.io/nzbget.Config>
 */
export interface NotiNzbgetConfig {
  url: string;
  user: string;
  pass: string;
};

/**
 * XmissionConfig is the Transmission input configuration.
 * @see golang: <github.com/Notifiarr/notifiarr/pkg/apps.XmissionConfig>
 */
export interface NotiXmissionConfig extends NotiExtraConfig {
  url: string;
  user: string;
  pass: string;
};

/**
 * @see golang: <github.com/Notifiarr/notifiarr/pkg/apps.TautulliConfig>
 */
export interface NotiTautulliConfig extends NotiExtraConfig, NotiTautulliConfig0 {};

/**
 * Config is the Tautulli configuration.
 * @see golang: <github.com/Notifiarr/notifiarr/pkg/apps/apppkg/tautulli.Config>
 */
export interface NotiTautulliConfig0 {
  url: string;
  apiKey: string;
};

/**
 * @see golang: <github.com/Notifiarr/notifiarr/pkg/apps.PlexConfig>
 */
export interface NotiPlexConfig extends NotiPlexConfig0, NotiExtraConfig {};

/**
 * @see golang: <github.com/Notifiarr/notifiarr/pkg/apps/apppkg/plex.Config>
 */
export interface NotiPlexConfig0 {
  url: string;
  token: string;
};

// Packages parsed:
//   1. github.com/Notifiarr/notifiarr/pkg/apps
//   2. github.com/Notifiarr/notifiarr/pkg/apps/apppkg/plex
//   3. github.com/Notifiarr/notifiarr/pkg/apps/apppkg/sabnzbd
//   4. github.com/Notifiarr/notifiarr/pkg/apps/apppkg/tautulli
//   5. github.com/Notifiarr/notifiarr/pkg/configfile
//   6. github.com/Notifiarr/notifiarr/pkg/logs
//   7. github.com/Notifiarr/notifiarr/pkg/services
//   8. github.com/Notifiarr/notifiarr/pkg/snapshot
//   9. github.com/Notifiarr/notifiarr/pkg/triggers/commands
//  10. github.com/Notifiarr/notifiarr/pkg/triggers/commands/cmdconfig
//  11. github.com/Notifiarr/notifiarr/pkg/triggers/common/scheduler
//  12. github.com/Notifiarr/notifiarr/pkg/triggers/endpoints/epconfig
//  13. github.com/Notifiarr/notifiarr/pkg/triggers/filewatch
//  14. golift.io/deluge
//  15. golift.io/nzbget
//  16. golift.io/qbit
//  17. golift.io/starr
