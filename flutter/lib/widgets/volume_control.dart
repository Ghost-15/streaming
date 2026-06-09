import 'package:flutter/material.dart';

class VolumeControl extends StatefulWidget {
  final double volume;
  final ValueChanged<double> onVolumeChanged;

  const VolumeControl({
    super.key,
    required this.volume,
    required this.onVolumeChanged,
  });

  @override
  State<VolumeControl> createState() => _VolumeControlState();
}

class _VolumeControlState extends State<VolumeControl> {
  late double _currentVolume;

  @override
  void initState() {
    super.initState();
    _currentVolume = widget.volume;
  }

  @override
  void didUpdateWidget(VolumeControl oldWidget) {
    super.didUpdateWidget(oldWidget);
    if (oldWidget.volume != widget.volume) {
      _currentVolume = widget.volume;
    }
  }

  @override
  Widget build(BuildContext context) {
    final colorScheme = Theme.of(context).colorScheme;
    final volumePercent = (_currentVolume * 100).toStringAsFixed(0);

    return Semantics(
      slider: true,
      label: 'Volume control',
      onIncrease: _currentVolume < 1.0
          ? () => setState(() {
                _currentVolume = (_currentVolume + 0.1).clamp(0.0, 1.0);
                widget.onVolumeChanged(_currentVolume);
              })
          : null,
      onDecrease: _currentVolume > 0.0
          ? () => setState(() {
                _currentVolume = (_currentVolume - 0.1).clamp(0.0, 1.0);
                widget.onVolumeChanged(_currentVolume);
              })
          : null,
      child: Column(
        crossAxisAlignment: CrossAxisAlignment.start,
        children: [
          Semantics(
            label: 'Volume: $volumePercent percent',
            child: Row(
              children: [
                Icon(Icons.volume_down, color: colorScheme.onSurfaceVariant),
                const SizedBox(width: 8),
                Expanded(
                  child: Slider(
                    value: _currentVolume,
                    min: 0,
                    max: 1,
                    divisions: 10,
                    label: '$volumePercent%',
                    onChanged: (value) {
                      setState(() => _currentVolume = value);
                      widget.onVolumeChanged(value);
                    },
                  ),
                ),
                const SizedBox(width: 8),
                Icon(Icons.volume_up, color: colorScheme.onSurfaceVariant),
              ],
            ),
          ),
          Padding(
            padding: const EdgeInsets.only(left: 8, top: 4),
            child: Text('Volume: $volumePercent%', style: Theme.of(context).textTheme.bodySmall),
          ),
        ],
      ),
    );
  }
}
