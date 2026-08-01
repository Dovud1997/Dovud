import 'package:flutter_riverpod/flutter_riverpod.dart';
import 'package:sfa_app/features/auth/data/auth_repository.dart';
import 'package:sfa_app/features/branding/data/branding_repository.dart';
import 'package:sfa_app/core/theme/brand_theme.dart';

class SessionState {
  const SessionState({
    this.session,
    this.branding,
    this.loading = true,
    this.error,
  });

  final AuthSession? session;
  final Branding? branding;
  final bool loading;
  final String? error;

  bool get isAuthenticated => session != null;

  SessionState copyWith({
    AuthSession? session,
    Branding? branding,
    bool? loading,
    String? error,
    bool clearSession = false,
    bool clearError = false,
  }) {
    return SessionState(
      session: clearSession ? null : (session ?? this.session),
      branding: branding ?? this.branding,
      loading: loading ?? this.loading,
      error: clearError ? null : (error ?? this.error),
    );
  }
}

class SessionController extends StateNotifier<SessionState> {
  SessionController(this._auth, this._branding) : super(const SessionState()) {
    bootstrap();
  }

  final AuthRepository _auth;
  final BrandingRepository _branding;

  Future<void> bootstrap() async {
    state = state.copyWith(loading: true, clearError: true);
    Branding branding;
    try {
      branding = await _branding.fetchPublic(tenantCode: 'demo');
    } catch (_) {
      branding = Branding.fallback();
    }
    final session = await _auth.restore();
    state = SessionState(session: session, branding: branding, loading: false);
  }

  Future<void> login({
    required String tenantCode,
    required String email,
    required String password,
  }) async {
    state = state.copyWith(loading: true, clearError: true);
    try {
      final branding = await _branding.fetchPublic(tenantCode: tenantCode);
      final session = await _auth.login(
        tenantCode: tenantCode,
        email: email,
        password: password,
      );
      state = SessionState(session: session, branding: branding, loading: false);
    } catch (e) {
      state = state.copyWith(loading: false, error: e.toString());
    }
  }

  Future<void> logout() async {
    await _auth.logout(state.session?.refreshToken);
    state = SessionState(branding: state.branding, loading: false);
  }
}

final sessionControllerProvider =
    StateNotifierProvider<SessionController, SessionState>((ref) {
  return SessionController(
    ref.watch(authRepositoryProvider),
    ref.watch(brandingRepositoryProvider),
  );
});
